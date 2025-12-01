from fastapi import FastAPI, Depends, Request, Form
from fastapi.responses import HTMLResponse, RedirectResponse, FileResponse, Response
from fastapi.templating import Jinja2Templates
from fastapi.staticfiles import StaticFiles
from sqlalchemy import create_engine, Column, Integer, String, ForeignKey, DateTime
from sqlalchemy.orm import sessionmaker, Session, relationship, joinedload
from sqlalchemy.ext.declarative import declarative_base
import datetime
import os
import io
from reportlab.pdfgen import canvas
from reportlab.lib.units import inch
from reportlab.lib.pagesizes import letter
from reportlab.lib.styles import getSampleStyleSheet
from reportlab.platypus import Paragraph, SimpleDocTemplate, Table, TableStyle, Spacer, PageBreak
from reportlab.lib import colors

# --- Configuration ---
DATABASE_URL = "sqlite:///./clefs.db"
Base = declarative_base()
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

from sqlalchemy import create_engine, Column, Integer, String, ForeignKey, DateTime, Table

# ... (keep the rest of the imports)

# --- Database Models ---

# Association table for the many-to-many relationship between Keys and Locations
key_location_association_table = Table('key_location_association', Base.metadata,
    Column('key_id', Integer, ForeignKey('keys.id')),
    Column('location_id', Integer, ForeignKey('locations.id'))
)

class Building(Base):
    __tablename__ = "buildings"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, unique=True, nullable=False)
    locations = relationship("Location", back_populates="building", cascade="all, delete-orphan")

class Location(Base):
    __tablename__ = "locations"
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    type = Column(String, nullable=False) # E.g., "Salle", "Porte", "Entrée"
    building_id = Column(Integer, ForeignKey("buildings.id"))
    
    building = relationship("Building", back_populates="locations")
    keys = relationship("Key", secondary=key_location_association_table, back_populates="locations")

class Key(Base):
    __tablename__ = "keys"
    id = Column(Integer, primary_key=True, index=True)
    number = Column(String, unique=True, index=True, nullable=False)
    description = Column(String)
    storage_location = Column(String, nullable=True)
    quantity_total = Column(Integer, default=1, nullable=False)
    quantity_reserve = Column(Integer, default=0, nullable=False)
    
    loans = relationship("Loan", back_populates="key", cascade="all, delete-orphan")
    locations = relationship("Location", secondary=key_location_association_table, back_populates="keys")

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

def create_summary_receipt_pdf(borrower: Borrower, loans: list[Loan]):
    """Generates a single PDF summarizing all active loans for a borrower."""
    buffer = io.BytesIO()
    c = canvas.Canvas(buffer, pagesize=letter)
    styles = getSampleStyleSheet()
    
    # Title
    c.setFont("Helvetica-Bold", 18)
    c.drawString(inch, 10 * inch, f"Récapitulatif des Emprunts")
    c.setFont("Helvetica-Bold", 14)
    c.drawString(inch, 9.7 * inch, f"Emprunteur : {borrower.name}")
    
    # Loan Details Table
    c.setFont("Helvetica", 10)
    text_y = 9 * inch
    line_height = 0.25 * inch

    # Table Header
    c.setFont("Helvetica-Bold", 10)
    c.drawString(inch, text_y, "N° Clé")
    c.drawString(2 * inch, text_y, "Description")
    c.drawString(5 * inch, text_y, "Date d'emprunt")
    c.line(inch, text_y - 0.1 * inch, 7.5 * inch, text_y - 0.1 * inch)
    text_y -= line_height

    # Table Rows
    c.setFont("Helvetica", 9)
    for loan in loans:
        c.drawString(inch, text_y, loan.key.number)
        c.drawString(2 * inch, text_y, loan.key.description)
        c.drawString(5 * inch, text_y, loan.loan_date.strftime('%d/%m/%Y %H:%M'))
        text_y -= line_height
        if text_y < 2 * inch: # Add new page if content is too long
            c.showPage()
            text_y = 10 * inch


    # Signature
    c.setFont("Helvetica", 12)
    c.drawString(inch, 1.5 * inch, "Signature de l'emprunteur :")
    c.line(3 * inch, 1.4 * inch, 6 * inch, 1.4 * inch)
    
    c.showPage()
    c.save()
    
    buffer.seek(0)
    return buffer

def create_key_plan_pdf(keys_data: list[Key], buildings_data: list[Building]):
    """
    Generates a comprehensive, multi-page PDF of the entire key plan using a simple, robust canvas-based approach.
    """
    buffer = io.BytesIO()
    c = canvas.Canvas(buffer, pagesize=letter)
    
    # Define layout constants
    margin = inch
    line_height = 0.25 * inch
    
    def check_page_break(y_pos):
        """Adds a new page if the y_pos is near the bottom margin."""
        if y_pos < margin:
            c.showPage()
            return 10.5 * inch # Reset y to top of page
        return y_pos

    # --- View by Key ---
    y = 10.5 * inch
    c.setFont("Helvetica-Bold", 16)
    c.drawString(margin, y, "Plan de Clés - Vue par Clé")
    y -= line_height * 2

    for key in keys_data:
        y = check_page_break(y)
        c.setFont("Helvetica-Bold", 12)
        c.drawString(margin, y, f"Clé : {key.number} ({key.description})")
        y -= line_height

        if not key.locations:
            c.setFont("Helvetica-Oblique", 10)
            c.drawString(margin + 0.2*inch, y, "N'ouvre aucun lieu.")
            y -= line_height
        else:
            c.setFont("Helvetica", 10)
            for loc in key.locations:
                y = check_page_break(y)
                building_name = loc.building.name if loc.building else "N/A"
                c.drawString(margin + 0.2*inch, y, f"- {building_name} &raquo; {loc.name}")
                y -= line_height
        y -= line_height * 0.5 # Extra space between keys

    # --- View by Location ---
    c.showPage()
    y = 10.5 * inch
    c.setFont("Helvetica-Bold", 16)
    c.drawString(margin, y, "Plan de Clés - Vue par Point d'Accès")
    y -= line_height * 2

    for building in buildings_data:
        y = check_page_break(y)
        c.setFont("Helvetica-Bold", 14)
        c.drawString(margin, y, f"Bâtiment : {building.name}")
        y -= line_height * 1.5

        if not building.locations:
            c.setFont("Helvetica-Oblique", 10)
            c.drawString(margin + 0.2*inch, y, "Aucun point d'accès défini.")
            y -= line_height
            continue

        for location in building.locations:
            y = check_page_break(y)
            c.setFont("Helvetica-Bold", 12)
            c.drawString(margin + 0.2*inch, y, f"Lieu : {location.name} ({location.type})")
            y -= line_height

            if not location.keys:
                c.setFont("Helvetica-Oblique", 10)
                c.drawString(margin + 0.4*inch, y, "Ouvert par aucune clé.")
                y -= line_height
            else:
                c.setFont("Helvetica", 10)
                keys_str = ", ".join([k.number for k in location.keys])
                c.drawString(margin + 0.4*inch, y, f"Clés : {keys_str}")
                y -= line_height
            y -= line_height * 0.5 # Extra space

    c.showPage()
    c.save()
    buffer.seek(0)
    return buffer
# --- FastAPI App ---

app = FastAPI(title="Gestionnaire de Clés")

# Mount static files
if not os.path.exists("app/static"):
    os.makedirs("app/static")
app.mount("/static", StaticFiles(directory="app/static"), name="static")

# Setup templates
if not os.path.exists("app/templates"):
    os.makedirs("app/templates")
templates = Jinja2Templates(directory="app/templates")

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
    active_loans = db.query(Loan).filter(Loan.return_date == None).all()

    # Create a more detailed loan info structure
    # For each key ID, we want a count and a list of borrower names
    loan_info = {}
    for loan in active_loans:
        if loan.key_id not in loan_info:
            loan_info[loan.key_id] = {'count': 0, 'borrowers': []}
        loan_info[loan.key_id]['count'] += 1
        loan_info[loan.key_id]['borrowers'].append(loan.borrower.name)
    
    return templates.TemplateResponse("index.html", {
        "request": request,
        "keys": keys,
        "loan_info": loan_info,
        "page_title": "Tableau de Bord"
    })

@app.get("/key-plan", response_class=HTMLResponse)
async def key_plan_page(request: Request, db: Session = Depends(get_db)):
    """
    Displays the key plan, showing relationships between keys and locations.
    """
    # Query for the "View by Key"
    keys_with_locations = db.query(Key).options(joinedload(Key.locations).joinedload(Location.building)).order_by(Key.number).all()
    
    # Query for the "View by Location" (grouped by building)
    buildings_with_data = db.query(Building).options(
        joinedload(Building.locations).joinedload(Location.keys)
    ).order_by(Building.name).all()

    return templates.TemplateResponse("key_plan.html", {
        "request": request,
        "keys_data": keys_with_locations,
        "buildings_data": buildings_with_data,
        "page_title": "Plan de Clés"
    })

@app.get("/key-plan/download")
async def download_key_plan_pdf(db: Session = Depends(get_db)):
    """
    Generates and returns a PDF of the entire key plan.
    """
    keys_with_locations = db.query(Key).options(joinedload(Key.locations).joinedload(Location.building)).order_by(Key.number).all()
    buildings_with_data = db.query(Building).options(
        joinedload(Building.locations).joinedload(Location.keys)
    ).order_by(Building.name).all()
    
    pdf_buffer = create_key_plan_pdf(keys_with_locations, buildings_with_data)
    
    filename = f"plan_de_cles_{datetime.date.today()}.pdf"
    headers = {'Content-Disposition': f'attachment; filename="{filename}"'}
    
    return Response(content=pdf_buffer.getvalue(), media_type='application/pdf', headers=headers)

@app.get("/manual", response_class=HTMLResponse)
async def manual_page(request: Request):
    """
    Displays the user manual page.
    """
    return templates.TemplateResponse("manual.html", {"request": request, "page_title": "Mode d'emploi"})

# --- Configuration Routes ---
@app.get("/config", response_class=HTMLResponse)
async def config_home(request: Request):
    return templates.TemplateResponse("config.html", {"request": request, "page_title": "Configuration"})

@app.get("/config/buildings", response_class=HTMLResponse)
async def manage_buildings(request: Request, db: Session = Depends(get_db)):
    buildings = db.query(Building).order_by(Building.name).all()
    return templates.TemplateResponse("buildings.html", {
        "request": request,
        "buildings": buildings,
        "page_title": "Gérer les Bâtiments"
    })

@app.post("/config/buildings/add", response_class=RedirectResponse)
async def add_building(db: Session = Depends(get_db), building_name: str = Form(...)):
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
async def manage_locations(request: Request, db: Session = Depends(get_db)):
    locations = db.query(Location).join(Building).order_by(Building.name, Location.name).all()
    buildings = db.query(Building).order_by(Building.name).all()
    return templates.TemplateResponse("locations.html", {
        "request": request,
        "locations": locations,
        "buildings": buildings,
        "page_title": "Gérer les Points d'Accès"
    })

@app.post("/config/locations/add", response_class=RedirectResponse)
async def add_location(
    db: Session = Depends(get_db),
    location_name: str = Form(...),
    location_type: str = Form(...),
    building_id: int = Form(...)
):
    new_location = Location(name=location_name, type=location_type, building_id=building_id)
    db.add(new_location)
    db.commit()
    return RedirectResponse(url="/config/locations", status_code=303)

@app.get("/config/locations/delete/{location_id}", response_class=RedirectResponse)
async def delete_location(location_id: int, db: Session = Depends(get_db)):
    location_to_delete = db.query(Location).filter(Location.id == location_id).first()
    if location_to_delete:
        db.delete(location_to_delete)
        db.commit()
    return RedirectResponse(url="/config/locations", status_code=303)


# --- Key Management Routes ---
@app.get("/keys", response_class=HTMLResponse)
async def manage_keys(request: Request, db: Session = Depends(get_db)):
    """
    Displays the page to manage keys, including all locations for the add form.
    """
    keys = db.query(Key).order_by(Key.number).all()
    buildings = db.query(Building).order_by(Building.name).all() # For the form
    
    # Get counts of active loans for each key type
    active_loans = db.query(Loan).filter(Loan.return_date == None).all()
    loan_counts = {}
    for loan in active_loans:
        loan_counts[loan.key_id] = loan_counts.get(loan.key_id, 0) + 1

    return templates.TemplateResponse("keys.html", {
        "request": request,
        "keys": keys,
        "loan_counts": loan_counts,
        "buildings": buildings,
        "page_title": "Gérer les Clés"
    })

@app.post("/keys/add", response_class=RedirectResponse)
async def add_key(
    db: Session = Depends(get_db),
    key_number: str = Form(...),
    key_description: str = Form(...),
    storage_location: str = Form(None),
    quantity_total: int = Form(1),
    quantity_reserve: int = Form(0),
    location_ids: list[int] = Form(None)
):
    """
    Handles the submission of the form to add a new key and associates it with locations.
    """
    new_key = Key(
        number=key_number, 
        description=key_description, 
        storage_location=storage_location,
        quantity_total=quantity_total,
        quantity_reserve=quantity_reserve
    )
    
    if location_ids:
        locations = db.query(Location).filter(Location.id.in_(location_ids)).all()
        new_key.locations = locations
        
    db.add(new_key)
    db.commit()
    return RedirectResponse(url="/keys", status_code=303)

@app.get("/keys/edit/{key_id}", response_class=HTMLResponse)
async def edit_key_form(request: Request, key_id: int, db: Session = Depends(get_db)):
    """
    Displays the form to edit an existing key.
    """
    key = db.query(Key).filter(Key.id == key_id).first()
    if not key:
        return HTMLResponse("Clé non trouvée", status_code=404)
    
    buildings = db.query(Building).order_by(Building.name).all()
    
    return templates.TemplateResponse("edit_key.html", {
        "request": request,
        "key": key,
        "buildings": buildings,
        "page_title": f"Modifier la Clé : {key.number}"
    })

@app.post("/keys/edit/{key_id}", response_class=RedirectResponse)
async def update_key(
    key_id: int,
    db: Session = Depends(get_db),
    key_number: str = Form(...),
    key_description: str = Form(...),
    storage_location: str = Form(None),
    quantity_total: int = Form(1),
    quantity_reserve: int = Form(0),
    location_ids: list[int] = Form(None)
):
    """
    Handles the submission of the form to update a key.
    """
    key_to_update = db.query(Key).filter(Key.id == key_id).first()
    if not key_to_update:
        return HTMLResponse("Clé non trouvée", status_code=404)

    # Update fields
    key_to_update.number = key_number
    key_to_update.description = key_description
    key_to_update.storage_location = storage_location
    key_to_update.quantity_total = quantity_total
    key_to_update.quantity_reserve = quantity_reserve
    
    # Update locations relationship
    if location_ids:
        locations = db.query(Location).filter(Location.id.in_(location_ids)).all()
        key_to_update.locations = locations
    else:
        key_to_update.locations = []
        
    db.commit()
    return RedirectResponse(url="/keys", status_code=303)


@app.get("/keys/delete/{key_id}", response_class=RedirectResponse)
async def delete_key(key_id: int, db: Session = Depends(get_db)):
    """
    Deletes a key by its ID.
    """
    key_to_delete = db.query(Key).filter(Key.id == key_id).first()
    if key_to_delete:
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
    # Treat empty email strings as NULL to avoid unique constraint violations
    email_to_store = borrower_email if borrower_email and borrower_email.strip() else None
    
    new_borrower = Borrower(name=borrower_name, email=email_to_store)
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
@app.get("/active-loans", response_class=HTMLResponse)
async def active_loans_page(request: Request, db: Session = Depends(get_db)):
    """
    Displays a page with all currently active (non-returned) loans,
    grouped by borrower.
    """
    active_loans = db.query(Loan).filter(Loan.return_date == None).options(
        joinedload(Loan.key),
        joinedload(Loan.borrower)
    ).order_by(Loan.borrower_id, Loan.loan_date.desc()).all()
    
    loans_by_borrower = {}
    for loan in active_loans:
        if loan.borrower not in loans_by_borrower:
            loans_by_borrower[loan.borrower] = []
        loans_by_borrower[loan.borrower].append(loan)

    return templates.TemplateResponse("active_loans.html", {
        "request": request,
        "loans_by_borrower": loans_by_borrower,
        "page_title": "Emprunts en Cours"
    })

@app.get("/loan/new", response_class=HTMLResponse)
async def new_loan_form(request: Request, key_id: int = None, db: Session = Depends(get_db)):
    """
    Displays the form to create a new loan, showing only available keys.
    """
    all_keys = db.query(Key).order_by(Key.number).all()
    active_loans = db.query(Loan).filter(Loan.return_date == None).all()
    
    loan_counts = {}
    for loan in active_loans:
        loan_counts[loan.key_id] = loan_counts.get(loan.key_id, 0) + 1
        
    available_keys = []
    for key in all_keys:
        loaned_count = loan_counts.get(key.id, 0)
        usable_quantity = key.quantity_total - key.quantity_reserve
        if usable_quantity > loaned_count:
            available_keys.append(key)

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
    key_ids: list[int] = Form(...),
    borrower_id: int = Form(...)
):
    """
    Creates multiple new loan records from a list of key IDs,
    after verifying availability for all keys.
    """
    # 1. Validation Phase: Check all keys before making any changes.
    for key_id in key_ids:
        key_to_loan = db.query(Key).filter(Key.id == key_id).first()
        if not key_to_loan:
            return HTMLResponse(f"Clé avec id {key_id} non trouvée", status_code=404)

        active_loan_count = db.query(Loan).filter(Loan.key_id == key_id, Loan.return_date == None).count()
        
        usable_quantity = key_to_loan.quantity_total - key_to_loan.quantity_reserve
        if active_loan_count >= usable_quantity:
            return HTMLResponse(f"La clé '{key_to_loan.number}' n'est plus en stock.", status_code=409)

    # 2. Creation Phase: If all keys are available, create all loans.
    loans_to_create = []
    for key_id in key_ids:
        new_loan_record = Loan(key_id=key_id, borrower_id=borrower_id, loan_date=datetime.datetime.utcnow())
        loans_to_create.append(new_loan_record)
    
    db.add_all(loans_to_create)
    db.commit()

    # For batch loans, we'll skip the single PDF receipt and go to the dashboard.
    return RedirectResponse(url="/", status_code=303)

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
    headers = {'Content-Disposition': f'attachment; filename="{filename}"'}
    
    return Response(content=pdf_buffer.getvalue(), media_type='application/pdf', headers=headers)

@app.get("/return/select/{key_id}", response_class=HTMLResponse)
async def select_loan_to_return(request: Request, key_id: int, db: Session = Depends(get_db)):
    """
    Displays a page to select a specific loan to return for a given key type.
    """
    key = db.query(Key).filter(Key.id == key_id).first()
    if not key:
        return HTMLResponse("Type de clé non trouvé", status_code=404)
    
    active_loans = db.query(Loan).filter(
        Loan.key_id == key_id,
        Loan.return_date == None
    ).order_by(Loan.loan_date.asc()).all()
    
    return templates.TemplateResponse("select_return.html", {
        "request": request,
        "key": key,
        "active_loans": active_loans,
        "page_title": "Choisir un Emprunt à Retourner"
    })

@app.get("/return/loan/{loan_id}", response_class=RedirectResponse)
async def return_loan(loan_id: int, db: Session = Depends(get_db)):
    """
    Marks a specific loan as returned by its ID.
    """
    loan_to_return = db.query(Loan).filter(Loan.id == loan_id).first()
    if loan_to_return:
        loan_to_return.return_date = datetime.datetime.utcnow()
        db.commit()
    # Redirect to the dashboard, or potentially back to the select page if it's still relevant
    return RedirectResponse(url="/", status_code=303)

@app.get("/loan/summary/borrower/{borrower_id}")
async def get_borrower_summary_pdf(borrower_id: int, db: Session = Depends(get_db)):
    """
    Generates and returns a summary PDF for all active loans of a specific borrower.
    """
    borrower = db.query(Borrower).filter(Borrower.id == borrower_id).first()
    if not borrower:
        return HTMLResponse("Emprunteur non trouvé", status_code=404)
    
    active_loans = db.query(Loan).filter(
        Loan.borrower_id == borrower_id,
        Loan.return_date == None
    ).options(joinedload(Loan.key)).order_by(Loan.loan_date.asc()).all()
    
    if not active_loans:
        return HTMLResponse("Cet emprunteur n'a aucun emprunt en cours.", status_code=404)

    pdf_buffer = create_summary_receipt_pdf(borrower, active_loans)
    
    filename = f"recapitulatif_{borrower.name.replace(' ', '_')}_{datetime.date.today()}.pdf"
    headers = {'Content-Disposition': f'attachment; filename="{filename}"'}
    
    return Response(content=pdf_buffer.getvalue(), media_type='application/pdf', headers=headers)


# --- Main execution ---
if __name__ == "__main__":
    import uvicorn
    # To run this app: uvicorn app.main:app --reload
    uvicorn.run(app, host="0.0.0.0", port=8000)
