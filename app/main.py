from fastapi import FastAPI, Depends, Request, Form
from fastapi.responses import HTMLResponse, RedirectResponse, StreamingResponse, JSONResponse
from fastapi.templating import Jinja2Templates
from fastapi.staticfiles import StaticFiles
from sqlalchemy import create_engine, Column, Integer, String, ForeignKey, DateTime, Table, func
from sqlalchemy.orm import sessionmaker, Session, relationship, joinedload
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
from typing import List, Optional
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

# Association table for many-to-many relationship between Key and Room
key_room_association = Table(
    'key_room_association', Base.metadata,
    Column('key_id', Integer, ForeignKey('keys.id'), primary_key=True),
    Column('room_id', Integer, ForeignKey('rooms.id'), primary_key=True)
)

# --- Database Models ---
class Key(Base):
    __tablename__ = "keys"
    id = Column(Integer, primary_key=True, index=True)
    number = Column(String, unique=True, index=True, nullable=False)
    description = Column(String)
    quantity_total = Column(Integer, default=1)
    quantity_reserve = Column(Integer, default=0)
    storage_location = Column(String, nullable=True)
    loans = relationship("Loan", back_populates="key", cascade="all, delete-orphan")
    rooms = relationship("Room", secondary=key_room_association, back_populates="keys")

class Room(Base):
    __tablename__ = "rooms"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, index=True, nullable=False)
    type = Column(String) # New column
    keys = relationship("Key", secondary=key_room_association, back_populates="rooms")
    building_id = Column(Integer, ForeignKey("buildings.id"))
    building = relationship("Building", back_populates="rooms")

class Borrower(Base):
    __tablename__ = "borrowers"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, index=True, nullable=False)
    email = Column(String, index=True)
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

class Building(Base):
    __tablename__ = "buildings"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, unique=True, index=True, nullable=False)
    rooms = relationship("Room", back_populates="building", cascade="all, delete-orphan")

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

def create_borrower_receipt_pdf(borrower: Borrower, loans: List[Loan]):
    """
    Generates a PDF with all keys borrowed by a specific borrower.
    """
    buffer = io.BytesIO()
    c = canvas.Canvas(buffer, pagesize=letter)
    styles = getSampleStyleSheet()
    
    # Title
    c.setFont("Helvetica-Bold", 18)
    c.drawString(inch, 10 * inch, "Bon de Sortie de Clés")
    
    # Borrower Details
    c.setFont("Helvetica", 12)
    text_y = 9.5 * inch
    
    c.drawString(inch, text_y, f"Emprunté par :")
    c.setFont("Helvetica-Bold", 12)
    c.drawString(3 * inch, text_y, f"{borrower.name}")
    
    c.setFont("Helvetica", 12)
    c.drawString(inch, text_y - 0.4 * inch, f"Date :")
    c.drawString(3 * inch, text_y - 0.4 * inch, f"{datetime.datetime.now().strftime('%d/%m/%Y à %H:%M')}")
    
    c.drawString(inch, text_y - 0.8 * inch, f"Nombre de clés :")
    c.drawString(3 * inch, text_y - 0.8 * inch, f"{len(loans)}")
    
    # Draw a line
    c.line(inch, text_y - 1 * inch, 7.5 * inch, text_y - 1 * inch)
    
    # List of keys
    c.setFont("Helvetica-Bold", 12)
    list_y = text_y - 1.5 * inch
    c.drawString(inch, list_y, "Liste des clés empruntées :")
    
    c.setFont("Helvetica", 11)
    list_y -= 0.4 * inch
    
    for i, loan in enumerate(loans, 1):
        if list_y < 2 * inch:  # Start a new page if running out of space
            c.showPage()
            c.setFont("Helvetica", 11)
            list_y = 10 * inch
        
        c.drawString(inch + 0.2 * inch, list_y, f"{i}.")
        c.drawString(inch + 0.5 * inch, list_y, f"{loan.key.number}")
        c.drawString(inch + 2 * inch, list_y, f"- {loan.key.description}")
        c.drawString(inch + 5 * inch, list_y, f"({loan.loan_date.strftime('%d/%m/%Y')})")
        list_y -= 0.3 * inch
    
    # Agreement Text
    list_y -= 0.5 * inch
    if list_y < 4 * inch:
        c.showPage()
        list_y = 9 * inch
    
    ptext = f"""
    Je soussigné(e), {borrower.name}, reconnais avoir reçu les {len(loans)} clé(s) mentionnée(s) ci-dessus.
    Je m'engage à en prendre soin et à les restituer à la fin de leur utilisation.
    En cas de perte ou de dégradation, je suis conscient(e) que ma responsabilité
    pourra être engagée.
    """
    
    p = Paragraph(ptext, styles['Normal'])
    p.wrapOn(c, 6 * inch, 5 * inch)
    p.drawOn(c, inch, list_y)
    
    # Signature
    list_y -= 2.5 * inch
    if list_y < 2 * inch:
        c.showPage()
        list_y = 8 * inch
    
    c.setFont("Helvetica", 12)
    c.drawString(inch, list_y, "Signature de l'emprunteur :")
    c.line(3 * inch, list_y - 0.1 * inch, 6 * inch, list_y - 0.1 * inch)
    
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
    active_loans = db.query(Loan).options(
        joinedload(Loan.borrower)
    ).filter(Loan.return_date == None).all()

    # Build loan_info: for each key ID provide count and borrower names
    loan_info = {}
    for loan in active_loans:
        lid = loan.key_id
        borrowers = loan_info.get(lid, SimpleNamespace(count=0, borrowers=[])).borrowers
        # increment count and append borrower
        if lid not in loan_info:
            loan_info[lid] = SimpleNamespace(count=0, borrowers=[])
        loan_info[lid].count += 1
        if loan.borrower:
            loan_info[lid].borrowers.append(loan.borrower.name)
    
    # Ensure template expects quantity_* attributes — provide sensible defaults
    for k in keys:
        if not hasattr(k, 'quantity_total'):
            setattr(k, 'quantity_total', 1)
        if not hasattr(k, 'quantity_reserve'):
            setattr(k, 'quantity_reserve', 0)

    return templates.TemplateResponse("index.html", {
        "request": request,
        "keys": keys,
        "loan_info": loan_info,
        "page_title": "Tableau de Bord"
    })

# --- Key Management Routes ---
@app.get("/keys", response_class=HTMLResponse)
async def manage_keys(request: Request, db: Session = Depends(get_db)):
    """
    Displays the page to manage keys (list and add form).
    """
    keys = db.query(Key).order_by(Key.number).all()
    # Eagerly load rooms for each building to use in the form
    buildings = db.query(Building).options(joinedload(Building.rooms).raiseload('*')).order_by(Building.name).all()
    
    # Correctly calculate active loan counts
    active_loans = db.query(Loan.key_id, func.count(Loan.key_id)).filter(Loan.return_date == None).group_by(Loan.key_id).all()
    loan_counts = {key_id: count for key_id, count in active_loans}

    return templates.TemplateResponse("keys.html", {
        "request": request,
        "keys": db.query(Key).options(
            joinedload(Key.rooms).joinedload(Room.building).raiseload('*'),
            joinedload(Key.rooms).raiseload('*')
        ).order_by(Key.number).all(),
        "buildings": buildings, # Pass buildings to the template
        "loan_counts": loan_counts, # Pass loan_counts to the template
        "page_title": "Gérer les Clés"
    })


@app.post("/keys/add", response_class=RedirectResponse)
async def add_key(
    db: Session = Depends(get_db),
    key_number: str = Form(...),
    key_description: str = Form(...),
    quantity_total: int = Form(1),
    quantity_reserve: int = Form(0),
    storage_location: str = Form(None),
    location_ids: Optional[List[int]] = Form(None) # Expect a list of room IDs
):
    """
    Handles the submission of the form to add a new key.
    """
    new_key = Key(
        number=key_number,
        description=key_description,
        quantity_total=quantity_total,
        quantity_reserve=quantity_reserve,
        storage_location=storage_location
    )
    db.add(new_key)
    db.flush() # Flush to get the new_key.id

    if location_ids:
        rooms = db.query(Room).filter(Room.id.in_(location_ids)).all()
        for room in rooms:
            new_key.rooms.append(room) # Associate rooms with the new key

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
    
    # Count active loans for each borrower
    active_loans = db.query(Loan.borrower_id, func.count(Loan.borrower_id)).filter(Loan.return_date == None).group_by(Loan.borrower_id).all()
    loan_counts = {borrower_id: count for borrower_id, count in active_loans}
    
    return templates.TemplateResponse("borrowers.html", {
        "request": request,
        "borrowers": borrowers,
        "loan_counts": loan_counts,
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

@app.get("/borrower/receipt/{borrower_id}")
async def get_borrower_receipt(borrower_id: int, db: Session = Depends(get_db)):
    """
    Generates and returns a PDF receipt with all keys borrowed by a specific borrower.
    """
    borrower = db.query(Borrower).filter(Borrower.id == borrower_id).first()
    if not borrower:
        return HTMLResponse("Emprunteur non trouvé", status_code=404)
    
    # Get all active loans for this borrower
    active_loans = db.query(Loan).options(
        joinedload(Loan.key)
    ).filter(Loan.borrower_id == borrower_id, Loan.return_date == None).all()
    
    if not active_loans:
        return HTMLResponse("Aucun emprunt actif pour cet emprunteur", status_code=404)
    
    pdf_buffer = create_borrower_receipt_pdf(borrower, active_loans)
    
    filename = f"bon_de_sortie_cles_{borrower.name.replace(' ', '_')}_{datetime.datetime.now().strftime('%Y%m%d')}.pdf"
    
    return StreamingResponse(pdf_buffer, media_type='application/pdf', headers={"Content-Disposition": f"attachment; filename=\"{filename}\""})


# --- Loan Management Routes ---
@app.get("/loan/new", response_class=HTMLResponse)
async def new_loan_form(request: Request, key_id: int = None, db: Session = Depends(get_db)):
    """
    Displays the form to create a new loan.
    """
    # Get all keys and count active loans for each
    all_keys = db.query(Key).order_by(Key.number).all()
    active_loans = db.query(Loan).filter(Loan.return_date == None).all()

    loan_counts = {}
    for loan in active_loans:
        loan_counts[loan.key_id] = loan_counts.get(loan.key_id, 0) + 1

    # A key is available if its total quantity is greater than the number of active loans
    available_keys = []
    for key in all_keys:
        usable_quantity = key.quantity_total - key.quantity_reserve
        if usable_quantity > loan_counts.get(key.id, 0):
            available_keys.append(key)

    borrowers = db.query(Borrower).order_by(Borrower.name).all()
    
    return templates.TemplateResponse("new_loan.html", {
        "request": request,
        "available_keys": available_keys,
        "borrowers": borrowers,
        "selected_key_id": key_id,
        "page_title": "Nouvel Emprunt"
    })

@app.post("/loan/new")
async def create_loan(
    db: Session = Depends(get_db),
    key_ids: List[int] = Form(...),
    borrower_id: int = Form(...)
):
    """
    Creates new loan records for each selected key.
    """
    # Get current loan counts for all keys
    active_loans = db.query(Loan.key_id, func.count(Loan.key_id)).filter(Loan.return_date == None).group_by(Loan.key_id).all()
    loan_counts = {key_id: count for key_id, count in active_loans}

    # Get the total quantity for each selected key
    keys_to_loan = db.query(Key).filter(Key.id.in_(key_ids)).all()
    key_quantities = {key.id: (key.quantity_total - key.quantity_reserve) for key in keys_to_loan}

    for key_id in key_ids:
        # Check if the key is available before creating the loan
        current_loans_for_key = loan_counts.get(key_id, 0)
        usable_quantity = key_quantities.get(key_id, 0)

        if usable_quantity > current_loans_for_key:
            new_loan_record = Loan(key_id=key_id, borrower_id=borrower_id, loan_date=datetime.datetime.utcnow())
            db.add(new_loan_record)
            loan_counts[key_id] = current_loans_for_key + 1 # Increment count for next check in the same request
    
    # After creating loans, check if only one was created to redirect to receipt
    new_loans = db.new
    single_new_loan = None
    if len(new_loans) == 1:
        single_new_loan = list(new_loans)[0]
    
    db.commit()
    if single_new_loan:
        return JSONResponse(content={"redirect_url": f"/loan/receipt/{single_new_loan.id}"})
    return JSONResponse(content={"redirect_url": "/active-loans"})

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
async def config_buildings(request: Request, db: Session = Depends(get_db)):
    buildings = db.query(Building).order_by(Building.name).all()
    return templates.TemplateResponse("buildings.html", {"request": request, "buildings": buildings, "page_title": "Gérer les Bâtiments"})


@app.post("/config/buildings/add", response_class=RedirectResponse)
async def add_building(
    db: Session = Depends(get_db),
    building_name: str = Form(...)
):
    """
    Handles the submission of the form to add a new building.
    """
    new_building = Building(name=building_name)
    db.add(new_building)
    db.commit()
    return RedirectResponse(url="/config/buildings", status_code=303)


@app.get("/config/buildings/delete/{building_id}", response_class=RedirectResponse)
async def delete_building(building_id: int, db: Session = Depends(get_db)):
    building_to_delete = db.query(Building).filter(Building.id == building_id).first()
    if building_to_delete:
        db.delete(building_to_delete)
        db.commit()
    return RedirectResponse(url="/config/buildings", status_code=303)


@app.get("/config/locations", response_class=HTMLResponse)
async def config_locations(request: Request, db: Session = Depends(get_db)):
    buildings = db.query(Building).options(joinedload(Building.rooms)).order_by(Building.name).all()
    rooms = db.query(Room).order_by(Room.name).all() # Fetch all rooms for the list
    return templates.TemplateResponse("locations.html", {
        "request": request,
        "buildings": buildings,
        "locations": rooms, # Pass rooms as locations to the template
        "page_title": "Gérer les Points d'Accès"
    })


@app.post("/config/locations/add", response_class=RedirectResponse)
async def add_location(
    db: Session = Depends(get_db),
    building_id: int = Form(...),
    location_name: str = Form(...),
    location_type: str = Form(...)
):
    """
    Handles the submission of the form to add a new room (location).
    """
    new_room = Room(
        name=location_name,
        type=location_type,
        building_id=building_id
    )
    db.add(new_room)
    db.commit()
    return RedirectResponse(url="/config/locations", status_code=303)


@app.get("/config/locations/delete/{location_id}", response_class=RedirectResponse)
async def delete_location(location_id: int, db: Session = Depends(get_db)):
    """
    Deletes a room (location) by its ID.
    """
    room_to_delete = db.query(Room).filter(Room.id == location_id).first()
    if room_to_delete:
        db.delete(room_to_delete)
        db.commit()
    return RedirectResponse(url="/config/locations", status_code=303)


@app.get("/active-loans", response_class=HTMLResponse)
async def view_active_loans(request: Request, db: Session = Depends(get_db)):
    # Build a grouping by borrower with loans
    active_loans = db.query(Loan).options(
        joinedload(Loan.key),
        joinedload(Loan.borrower)
    ).filter(Loan.return_date == None).order_by(Loan.borrower_id, Loan.loan_date).all()
    
    loans_by_borrower = {}
    for loan in active_loans:
        # Group by the borrower object itself to handle multiple loans correctly
        loans_by_borrower.setdefault(loan.borrower, []).append(loan)

    return templates.TemplateResponse("active_loans.html", {"request": request, "loans_by_borrower": loans_by_borrower, "page_title": "Emprunts en Cours"})

@app.get("/loans/report", response_class=HTMLResponse)
async def loans_report(request: Request, db: Session = Depends(get_db)):
    """
    Displays a complete report of all active loans (keys currently out).
    """
    active_loans = db.query(Loan).options(
        joinedload(Loan.key),
        joinedload(Loan.borrower)
    ).filter(Loan.return_date == None).order_by(Loan.key_id, Loan.loan_date).all()
    
    return templates.TemplateResponse("loans_report.html", {
        "request": request,
        "active_loans": active_loans,
        "now": datetime.datetime.now(),
        "page_title": "Rapport des Clés Sorties"
    })


@app.get("/key-plan", response_class=HTMLResponse)
async def view_key_plan(request: Request, db: Session = Depends(get_db)):
    # Fetch keys with their associated rooms
    keys = db.query(Key).options(joinedload(Key.rooms).joinedload(Room.building)).order_by(Key.number).all()

    # Fetch buildings with their associated rooms and the keys that open those rooms
    buildings = db.query(Building).options(
        joinedload(Building.rooms).joinedload(Room.keys)
    ).order_by(Building.name).all()
    
    return templates.TemplateResponse("key_plan.html", {
        "request": request,
        "keys_data": keys,
        "buildings_data": buildings,
        "page_title": "Plan de Clés"
    })


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
    buildings = db.query(Building).options(joinedload(Building.rooms)).order_by(Building.name).all()
    return templates.TemplateResponse("edit_key.html", {
        "request": request, "key": key, "buildings": buildings,
        "page_title": f"Modifier la Clé {key.number}"
    })

@app.post("/keys/edit/{key_id}", response_class=RedirectResponse)
async def edit_key_submit(
    key_id: int,
    db: Session = Depends(get_db),
    key_number: str = Form(...),
    key_description: str = Form(...),
    quantity_total: int = Form(...),
    quantity_reserve: int = Form(...),
    storage_location: str = Form(None),
    location_ids: Optional[List[int]] = Form(None)
):
    key = db.query(Key).filter(Key.id == key_id).first()
    if key:
        key.number = key_number
        key.description = key_description
        key.quantity_total = quantity_total
        key.quantity_reserve = quantity_reserve
        key.storage_location = storage_location
        
        # Update associated rooms
        key.rooms.clear()
        if location_ids:
            rooms = db.query(Room).filter(Room.id.in_(location_ids)).all()
            key.rooms.extend(rooms)

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


@app.get("/about", response_class=HTMLResponse)
async def about_page(request: Request):
    return templates.TemplateResponse("about.html", {"request": request, "page_title": "À Propos"})


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
