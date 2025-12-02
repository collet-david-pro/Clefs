from fastapi import FastAPI, Depends, Request, Form
from fastapi.responses import HTMLResponse, RedirectResponse, StreamingResponse
from fastapi.templating import Jinja2Templates
from fastapi.staticfiles import StaticFiles
from sqlalchemy import create_engine, Column, Integer, String, ForeignKey, DateTime
from sqlalchemy.orm import sessionmaker, Session, relationship
from sqlalchemy.ext.declarative import declarative_base
import datetime
import os
import sys
import threading
import sys
import webbrowser
import io
from reportlab.pdfgen import canvas
from reportlab.lib.units import inch
from reportlab.lib.pagesizes import letter
from reportlab.lib.styles import getSampleStyleSheet
from reportlab.platypus import Paragraph
from types import SimpleNamespace
import uvicorn
import webview

# --- Helper for PyInstaller ---
def resource_path(relative_path):
    """ Get absolute path to resource, works for dev and for PyInstaller """
    try:
        # PyInstaller creates a temp folder and stores path in _MEIPASS
        base_path = sys._MEIPASS
    except Exception:
        base_path = os.path.abspath(".")
    return os.path.join(base_path, relative_path)
# --- Configuration ---
DATABASE_URL = "sqlite:///./clefs.db"
Base = declarative_base()
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# --- Database Models ---
class Key(Base):
    __tablename__ = "keys"
    id = Column(Integer, primary_key=True, index=True)
    number = Column(String, unique=True, index=True, nullable=False)
    description = Column(String)
    loans = relationship("Loan", back_populates="key", cascade="all, delete-orphan")
    rooms = relationship("Room", back_populates="key", cascade="all, delete-orphan")

class Room(Base):
    __tablename__ = "rooms"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, unique=True, index=True, nullable=False)
    key_id = Column(Integer, ForeignKey("keys.id"))
    key = relationship("Key", back_populates="rooms")

class Borrower(Base):
    __tablename__ = "borrowers"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, unique=True, index=True, nullable=False)
    email = Column(String, unique=True, index=True)
    loans = relationship("Loan", back_populates="borrower", cascade="all, delete-orphan")

class Loan(Base):
    __tablename__ = "loans"
    id = Column(Integer, primary_key=True, index=True)
    key_id = Column(Integer, ForeignKey("keys.id"), nullable=False)
    borrower_id = Column(Integer, ForeignKey("borrowers.id"), nullable=False)
    loan_date = Column(DateTime, default=datetime.datetime.utcnow)
    return_date = Column(DateTime, nullable=True)
    
    key = relationship("Key", back_populates="loans")
    borrower = relationship("Borrower", back_populates="loans")

# Create database tables
Base.metadata.create_all(bind=engine)

# --- PDF Generation ---
def create_receipt_pdf(loan: Loan):
    buffer = io.BytesIO()
    c = canvas.Canvas(buffer, pagesize=letter)
    styles = getSampleStyleSheet()
    
    # Title
    c.setFont("Helvetica-Bold", 18)
    c.drawString(inch, 10 * inch, "Bon de Sortie de Clé")
    
    # Loan Details
    c.setFont("Helvetica", 12)
    text_y = 9 * inch
    
    c.drawString(inch, text_y, f"Numéro de la clé :")
    c.drawString(3 * inch, text_y, f"{loan.key.number}")
    
    c.drawString(inch, text_y - 0.5 * inch, f"Description :")
    c.drawString(3 * inch, text_y - 0.5 * inch, f"{loan.key.description}")

    c.drawString(inch, text_y - 1 * inch, f"Emprunté par :")
    c.drawString(3 * inch, text_y - 1 * inch, f"{loan.borrower.name}")
    
    c.drawString(inch, text_y - 1.5 * inch, f"Date d'emprunt :")
    c.drawString(3 * inch, text_y - 1.5 * inch, f"{loan.loan_date.strftime('%d/%m/%Y à %H:%M')}")

    # Agreement Text
    story = []
    ptext = """
    Je soussigné(e), {}, reconnais avoir reçu la clé mentionnée ci-dessus.
    Je m'engage à en prendre soin et à la restituer à la fin de son utilisation.
    En cas de perte ou de dégradation, je suis conscient(e) que ma responsabilité
    pourra être engagée.
    """.format(loan.borrower.name)
    story.append(Paragraph(ptext, styles['Normal']))
    
    p = Paragraph(ptext, styles['Normal'])
    p.wrapOn(c, 6 * inch, 5 * inch)
    p.drawOn(c, inch, 6.5 * inch)


    # Signature
    c.drawString(inch, 4 * inch, "Signature de l'emprunteur :")
    c.line(3 * inch, 3.9 * inch, 6 * inch, 3.9 * inch)

    c.showPage()
    c.save()
    
    buffer.seek(0)
    return buffer

# --- FastAPI App ---
app = FastAPI(title="Gestionnaire de Clés")

# Mount static files
if not os.path.exists("app/static"):
    os.makedirs("app/static") # This might not be needed if static dir is always present
app.mount("/static", StaticFiles(directory=resource_path("app/static")), name="static")

# Setup templates
if not os.path.exists("app/templates"):
    os.makedirs("app/templates")
templates = Jinja2Templates(directory=resource_path("app/templates"))

# Dependency to get DB session
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

# --- Routes ---
@app.get("/", response_class=HTMLResponse)
async def read_root(request: Request, db: Session = Depends(get_db)):
    """
    Homepage: Displays the current status of all keys.
    """
    keys = db.query(Key).order_by(Key.number).all()
    
    # Find active loans (not returned yet)
    active_loans = db.query(Loan).filter(Loan.return_date == None).all()
    
    # Create a dictionary to easily find who has which key
    loan_map = {loan.key_id: loan.borrower.name for loan in active_loans}
    
    return templates.TemplateResponse("index.html", {
        "request": request,
        "keys": keys,
        "loan_map": loan_map,
        "page_title": "Tableau de Bord"
    })

# --- Key Management Routes ---
@app.get("/keys", response_class=HTMLResponse)
async def manage_keys(request: Request, db: Session = Depends(get_db)):
    """
    Displays the page to manage keys (list and add form).
    """
    keys = db.query(Key).order_by(Key.number).all()
    return templates.TemplateResponse("keys.html", {
        "request": request,
        "keys": keys,
        "page_title": "Gérer les Clés"
    })

@app.post("/keys/add", response_class=RedirectResponse)
async def add_key(
    db: Session = Depends(get_db),
    key_number: str = Form(...),
    key_description: str = Form(...)
):
    """
    Handles the submission of the form to add a new key.
    """
    new_key = Key(number=key_number, description=key_description)
    db.add(new_key)
    db.commit()
    return RedirectResponse(url="/keys", status_code=303)

@app.get("/keys/delete/{key_id}", response_class=RedirectResponse)
async def delete_key(key_id: int, db: Session = Depends(get_db)):
    """
    Deletes a key by its ID.
    """
    key_to_delete = db.query(Key).filter(Key.id == key_id).first()
    if key_to_delete:
        # Before deleting a key, you might want to ensure it has no active loans
        # or re-associate rooms. For now, we'll just delete it.
        db.delete(key_to_delete)
        db.commit()
    return RedirectResponse(url="/keys", status_code=303)


# --- Borrower Management Routes ---
@app.get("/borrowers", response_class=HTMLResponse)
async def manage_borrowers(request: Request, db: Session = Depends(get_db)):
    """
    Displays the page to manage borrowers (list and add form).
    """
    borrowers = db.query(Borrower).order_by(Borrower.name).all()
    return templates.TemplateResponse("borrowers.html", {
        "request": request,
        "borrowers": borrowers,
        "page_title": "Gérer les Emprunteurs"
    })

@app.post("/borrowers/add", response_class=RedirectResponse)
async def add_borrower(
    db: Session = Depends(get_db),
    borrower_name: str = Form(...),
    borrower_email: str = Form(None)
):
    """
    Handles the submission of the form to add a new borrower.
    """
    new_borrower = Borrower(name=borrower_name, email=borrower_email)
    db.add(new_borrower)
    db.commit()
    return RedirectResponse(url="/borrowers", status_code=303)

@app.get("/borrowers/delete/{borrower_id}", response_class=RedirectResponse)
async def delete_borrower(borrower_id: int, db: Session = Depends(get_db)):
    """
    Deletes a borrower by their ID.
    """
    borrower_to_delete = db.query(Borrower).filter(Borrower.id == borrower_id).first()
    if borrower_to_delete:
        db.delete(borrower_to_delete)
        db.commit()
    return RedirectResponse(url="/borrowers", status_code=303)


# --- Loan Management Routes ---
@app.get("/loan/new", response_class=HTMLResponse)
async def new_loan_form(request: Request, key_id: int = None, db: Session = Depends(get_db)):
    """
    Displays the form to create a new loan.
    """
    # Get keys that are not currently on loan
    active_loan_key_ids = [loan.key_id for loan in db.query(Loan).filter(Loan.return_date == None).all()]
    available_keys = db.query(Key).filter(Key.id.notin_(active_loan_key_ids)).order_by(Key.number).all()
    
    borrowers = db.query(Borrower).order_by(Borrower.name).all()
    
    return templates.TemplateResponse("new_loan.html", {
        "request": request,
        "available_keys": available_keys,
        "borrowers": borrowers,
        "selected_key_id": key_id,
        "page_title": "Nouvel Emprunt"
    })

@app.post("/loan/new", response_class=RedirectResponse)
async def create_loan(
    db: Session = Depends(get_db),
    key_id: int = Form(...),
    borrower_id: int = Form(...)
):
    """
    Creates a new loan record and redirects to the PDF receipt.
    """
    # Double-check the key is actually available
    existing_loan = db.query(Loan).filter(Loan.key_id == key_id, Loan.return_date == None).first()
    if existing_loan:
        # This should ideally return an error message to the user
        return RedirectResponse(url="/", status_code=303)

    new_loan_record = Loan(key_id=key_id, borrower_id=borrower_id, loan_date=datetime.datetime.utcnow())
    db.add(new_loan_record)
    db.commit()
    db.refresh(new_loan_record) # To get the new loan's ID

    return RedirectResponse(url=f"/loan/receipt/{new_loan_record.id}", status_code=303)

@app.get("/loan/receipt/{loan_id}")
async def get_loan_receipt(loan_id: int, db: Session = Depends(get_db)):
    """
    Generates and returns a PDF receipt for a specific loan.
    """
    loan = db.query(Loan).filter(Loan.id == loan_id).first()
    if not loan:
        return HTMLResponse("Emprunt non trouvé", status_code=404)
    
    pdf_buffer = create_receipt_pdf(loan)
    
    filename = f"bon_de_sortie_cle_{loan.key.number}_{loan.loan_date.strftime('%Y%m%d')}.pdf"
    
    # Return the generated PDF as a streaming response (BytesIO)
    return StreamingResponse(pdf_buffer, media_type='application/pdf', headers={"Content-Disposition": f"attachment; filename=\"{filename}\""})

@app.get("/return/{key_id}", response_class=RedirectResponse)
async def return_key(key_id: int, db: Session = Depends(get_db)):
    """
    Marks a key as returned by setting the return_date on the active loan.
    """
    active_loan = db.query(Loan).filter(Loan.key_id == key_id, Loan.return_date == None).first()
    if active_loan:
        active_loan.return_date = datetime.datetime.utcnow()
        db.commit()
    return RedirectResponse(url="/", status_code=303)


# --- Extra pages/routes to match templates (placeholders to avoid 404s) ---
@app.get("/config", response_class=HTMLResponse)
async def config_page(request: Request):
    return templates.TemplateResponse("config.html", {"request": request, "page_title": "Configuration"})


@app.get("/config/buildings", response_class=HTMLResponse)
async def config_buildings(request: Request):
    # Minimal implementation: no DB model for Building yet -> show empty list
    buildings = []
    return templates.TemplateResponse("buildings.html", {"request": request, "buildings": buildings, "page_title": "Gérer les Bâtiments"})


@app.post("/config/buildings/add", response_class=RedirectResponse)
async def add_building(request: Request, building_name: str = Form(...)):
    # Placeholder behaviour: do nothing persistent yet
    return RedirectResponse(url="/config/buildings", status_code=303)


@app.get("/config/buildings/delete/{building_id}", response_class=RedirectResponse)
async def delete_building(building_id: int):
    # Placeholder: no-op
    return RedirectResponse(url="/config/buildings", status_code=303)


@app.get("/config/locations", response_class=HTMLResponse)
async def config_locations(request: Request, db: Session = Depends(get_db)):
    # Minimal implementation: return empty lists when Location model not implemented
    buildings = []
    locations = []
    return templates.TemplateResponse("locations.html", {"request": request, "buildings": buildings, "locations": locations, "page_title": "Gérer les Points d'Accès"})


@app.post("/config/locations/add", response_class=RedirectResponse)
async def add_location(request: Request, building_id: int = Form(...), location_name: str = Form(...), location_type: str = Form(...)):
    # Placeholder: no-op
    return RedirectResponse(url="/config/locations", status_code=303)


@app.get("/config/locations/delete/{location_id}", response_class=RedirectResponse)
async def delete_location(location_id: int):
    # Placeholder: no-op
    return RedirectResponse(url="/config/locations", status_code=303)


@app.get("/active-loans", response_class=HTMLResponse)
async def view_active_loans(request: Request, db: Session = Depends(get_db)):
    # Build a grouping by borrower with loans
    active_loans = db.query(Loan).filter(Loan.return_date == None).all()
    loans_by_borrower = {}
    for loan in active_loans:
        name = loan.borrower.name if loan.borrower else "Inconnu"
        loans_by_borrower.setdefault(name, []).append(loan)

    return templates.TemplateResponse("active_loans.html", {"request": request, "loans_by_borrower": loans_by_borrower, "page_title": "Emprunts en Cours"})


@app.get("/loan/summary/borrower/{borrower_id}", response_class=HTMLResponse)
async def loan_summary_borrower(request: Request, borrower_id: int, db: Session = Depends(get_db)):
    borrower = db.query(Borrower).filter(Borrower.id == borrower_id).first()
    if not borrower:
        return HTMLResponse("Emprunteur non trouvé", status_code=404)

    loans = db.query(Loan).filter(Loan.borrower_id == borrower_id).all()
    # Reuse the active_loans template: build dict with a single borrower
    loans_by_borrower = {borrower.name: loans}
    return templates.TemplateResponse("active_loans.html", {"request": request, "loans_by_borrower": loans_by_borrower, "page_title": f"Emprunts de {borrower.name}"})


@app.get("/key-plan", response_class=HTMLResponse)
async def view_key_plan(request: Request, db: Session = Depends(get_db)):
    # Build minimal structures for the template — don't assume Location/Building models exist
    keys = db.query(Key).order_by(Key.number).all()
    keys_data = []
    for k in keys:
        # Provide an empty locations list to avoid template errors
        keys_data.append({"number": k.number, "locations": []})

    buildings_data = []
    return templates.TemplateResponse("key_plan.html", {"request": request, "keys_data": keys_data, "buildings_data": buildings_data, "page_title": "Plan de Clés"})


@app.get("/key-plan/download", response_class=HTMLResponse)
async def download_key_plan(request: Request):
    # Placeholder: just redirect back to the key plan page for now
    return RedirectResponse(url="/key-plan", status_code=303)


@app.get("/manual", response_class=HTMLResponse)
async def manual_page(request: Request):
    return templates.TemplateResponse("manual.html", {"request": request, "page_title": "Mode d\'emploi"})


@app.get("/keys/edit/{key_id}", response_class=HTMLResponse)
async def edit_key_form(request: Request, key_id: int, db: Session = Depends(get_db)):
    key = db.query(Key).filter(Key.id == key_id).first()
    if not key:
        return HTMLResponse("Clé non trouvée", status_code=404)
    return templates.TemplateResponse("edit_key.html", {"request": request, "key": key, "page_title": f"Modifier la Clé {key.number}"})


@app.post("/keys/edit/{key_id}", response_class=RedirectResponse)
async def edit_key_submit(key_id: int, db: Session = Depends(get_db), key_number: str = Form(...), key_description: str = Form(None)):
    key = db.query(Key).filter(Key.id == key_id).first()
    if key:
        key.number = key_number
        key.description = key_description
        db.commit()
    return RedirectResponse(url="/keys", status_code=303)


@app.get("/return/select/{key_id}", response_class=HTMLResponse)
async def select_return(request: Request, key_id: int, db: Session = Depends(get_db)):
    key = db.query(Key).filter(Key.id == key_id).first()
    if not key:
        return HTMLResponse("Clé non trouvée", status_code=404)

    active_loans = db.query(Loan).filter(Loan.key_id == key_id, Loan.return_date == None).all()
    return templates.TemplateResponse("select_return.html", {"request": request, "key": key, "active_loans": active_loans, "page_title": "Sélectionner le Retour"})


@app.get("/return/loan/{loan_id}", response_class=RedirectResponse)
async def return_loan(loan_id: int, db: Session = Depends(get_db)):
    loan = db.query(Loan).filter(Loan.id == loan_id).first()
    if loan and loan.return_date is None:
        loan.return_date = datetime.datetime.utcnow()
        db.commit()
    return RedirectResponse(url="/", status_code=303)

# --- Desktop App Integration (pywebview) ---

def run_server():
    """Function to run Uvicorn server."""
    uvicorn.run(app, host="127.0.0.1", port=8000, log_level="warning")

# --- Main execution ---
if __name__ == "__main__":
    # Si le script est lancé en tant qu'exécutable compilé par PyInstaller...
    if getattr(sys, 'frozen', False):
        # Lancer le serveur dans un thread séparé
        server_thread = threading.Thread(target=run_server)
        server_thread.daemon = True
        server_thread.start()

        # Créer la fenêtre de l'application de bureau
        webview.create_window(
            'Gestionnaire de Clés',
            'http://127.0.0.1:8000',
            width=1280,
            height=800
        )
        webview.start()
    else:
        # Comportement normal pour le développement (avec --reload)
        uvicorn.run("app.main:app", host="127.0.0.1", port=8000, reload=True)
