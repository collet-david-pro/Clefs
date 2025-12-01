import random
from sqlalchemy.orm import sessionmaker, joinedload
from app.main import Base, engine, Building, Location, Key, Borrower, Loan

# IMPORTANT: Run this script while the main FastAPI application is NOT running.

# Setup session
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
db = SessionLocal()

def clear_data():
    """Clears all data from the tables in the correct order."""
    print("--- Clearing all existing data... ---")
    # The association table is handled automatically by SQLAlchemy's ORM
    db.query(Loan).delete(synchronize_session=False)
    db.query(Key).delete(synchronize_session=False)
    db.query(Location).delete(synchronize_session=False)
    db.query(Building).delete(synchronize_session=False)
    db.query(Borrower).delete(synchronize_session=False)
    db.commit()
    print("Done.")

def seed_data():
    """Fills the database with a set of test data."""
    print("--- Seeding new data... ---")

    # --- 1. Create Buildings ---
    building_names = ["Bâtiment Central", "Annexe Ouest", "Pôle Technique Est"]
    buildings = [Building(name=name) for name in building_names]
    db.add_all(buildings)
    db.commit()
    print(f"Created {len(buildings)} buildings.")

    # --- 2. Create Locations ---
    locations = []
    location_types = ["Salle", "Porte", "Couloir", "Bureau", "Archive", "Entrée"]
    for building in buildings:
        for i in range(1, 16):
            loc_type = random.choice(location_types)
            floor = i // 5 + 1
            room_num = f"{floor}0{i % 5}" if i % 5 > 0 else f"{floor-1}05"
            locations.append(Location(
                building_id=building.id,
                name=f"{loc_type} {room_num}",
                type=loc_type
            ))
    db.add_all(locations)
    db.commit()
    print(f"Created {len(locations)} locations.")

    # --- 3. Create Borrowers ---
    borrowers = [Borrower(name=f"Emprunteur {i}", email=f"user{i}@example.com") for i in range(1, 21)]
    db.add_all(borrowers)
    db.commit()
    print(f"Created {len(borrowers)} borrowers.")

    # --- 4. Create Keys ---
    passe_types = ["Passe Partiel", "Passe Général", "Clé unique"]
    storage_locations = ["Accueil", "Administration", "Réserve"]
    all_locations = db.query(Location).all()
    keys = []
    for i in range(1, 11):
        key_type = random.choice(passe_types)
        total = random.randint(2, 10)
        reserve = random.randint(0, total // 2) # Reserve is at most half of total
        key = Key(
            number=f"K{i}-{key_type[0]}P",
            description=f"{key_type} N°{i}",
            storage_location=random.choice(storage_locations),
            quantity_total=total,
            quantity_reserve=reserve
        )
        # Assign a random number of locations to this key
        num_locations_for_key = 1 if key_type == "Clé unique" else random.randint(2, 10)
        key.locations = random.sample(all_locations, num_locations_for_key)
        keys.append(key)
    
    db.add_all(keys)
    db.commit()
    print(f"Created {len(keys)} key types.")
    
    print("--- Seeding complete. ---")


if __name__ == "__main__":
    try:
        clear_data()
        seed_data()
    except Exception as e:
        print(f"\nAn error occurred: {e}")
        db.rollback()
    finally:
        db.close()
        print("Session closed.")
