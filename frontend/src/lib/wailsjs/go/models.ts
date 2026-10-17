export namespace app {
	
	export class AccessDTO {
	    id: number;
	    name: string;
	    type: string;
	    buildingId: number;
	    buildingName: string;
	    floor: string;
	    category: string;
	    notes: string;
	    keyIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new AccessDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.buildingId = source["buildingId"];
	        this.buildingName = source["buildingName"];
	        this.floor = source["floor"];
	        this.category = source["category"];
	        this.notes = source["notes"];
	        this.keyIds = source["keyIds"];
	    }
	}
	export class HolderDTO {
	    borrowerId: number;
	    borrowerName: string;
	    keyNumber: string;
	
	    static createFrom(source: any = {}) {
	        return new HolderDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.borrowerId = source["borrowerId"];
	        this.borrowerName = source["borrowerName"];
	        this.keyNumber = source["keyNumber"];
	    }
	}
	export class KeyDTO {
	    id: number;
	    number: string;
	    description: string;
	    category: string;
	    storageLocation: string;
	    notes: string;
	    total: number;
	    reserve: number;
	    loaned: number;
	    available: number;
	    holders: string[];
	    accessIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new KeyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.number = source["number"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.storageLocation = source["storageLocation"];
	        this.notes = source["notes"];
	        this.total = source["total"];
	        this.reserve = source["reserve"];
	        this.loaned = source["loaned"];
	        this.available = source["available"];
	        this.holders = source["holders"];
	        this.accessIds = source["accessIds"];
	    }
	}
	export class AccessDetail {
	    access: AccessDTO;
	    keys: KeyDTO[];
	    holders: HolderDTO[];
	
	    static createFrom(source: any = {}) {
	        return new AccessDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.access = this.convertValues(source["access"], AccessDTO);
	        this.keys = this.convertValues(source["keys"], KeyDTO);
	        this.holders = this.convertValues(source["holders"], HolderDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AccessInput {
	    id: number;
	    name: string;
	    type: string;
	    buildingId: number;
	    floor: string;
	    category: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new AccessInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.buildingId = source["buildingId"];
	        this.floor = source["floor"];
	        this.category = source["category"];
	        this.notes = source["notes"];
	    }
	}
	export class AnomalyDTO {
	    keyId: number;
	    keyNumber: string;
	    total: number;
	    reserve: number;
	    loaned: number;
	    available: number;
	
	    static createFrom(source: any = {}) {
	        return new AnomalyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keyId = source["keyId"];
	        this.keyNumber = source["keyNumber"];
	        this.total = source["total"];
	        this.reserve = source["reserve"];
	        this.loaned = source["loaned"];
	        this.available = source["available"];
	    }
	}
	export class AppInfo {
	    version: string;
	    dataDir: string;
	    dbPath: string;
	    documentsDir: string;
	    backupsDir: string;
	    agent: string;
	    isEmpty: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.dataDir = source["dataDir"];
	        this.dbPath = source["dbPath"];
	        this.documentsDir = source["documentsDir"];
	        this.backupsDir = source["backupsDir"];
	        this.agent = source["agent"];
	        this.isEmpty = source["isEmpty"];
	    }
	}
	export class BackupDTO {
	    path: string;
	    name: string;
	    size: string;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.date = source["date"];
	    }
	}
	export class BorrowerDTO {
	    id: number;
	    name: string;
	    email: string;
	    status: string;
	    phone: string;
	    activeLoans: number;
	
	    static createFrom(source: any = {}) {
	        return new BorrowerDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.status = source["status"];
	        this.phone = source["phone"];
	        this.activeLoans = source["activeLoans"];
	    }
	}
	export class LoanDTO {
	    id: number;
	    keyId: number;
	    keyNumber: string;
	    keyDescription: string;
	    borrowerId: number;
	    borrowerName: string;
	    loanDate: string;
	    returnDate: string;
	    plannedReturnDate: string;
	    loanType: string;
	    condition: string;
	    createdBy: string;
	    status: string;
	    days: number;
	
	    static createFrom(source: any = {}) {
	        return new LoanDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.keyId = source["keyId"];
	        this.keyNumber = source["keyNumber"];
	        this.keyDescription = source["keyDescription"];
	        this.borrowerId = source["borrowerId"];
	        this.borrowerName = source["borrowerName"];
	        this.loanDate = source["loanDate"];
	        this.returnDate = source["returnDate"];
	        this.plannedReturnDate = source["plannedReturnDate"];
	        this.loanType = source["loanType"];
	        this.condition = source["condition"];
	        this.createdBy = source["createdBy"];
	        this.status = source["status"];
	        this.days = source["days"];
	    }
	}
	export class BorrowerDetail {
	    borrower: BorrowerDTO;
	    activeLoans: LoanDTO[];
	    coveredAccesses: AccessDTO[];
	    redundancies: AccessDTO[];
	    history: LoanDTO[];
	
	    static createFrom(source: any = {}) {
	        return new BorrowerDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.borrower = this.convertValues(source["borrower"], BorrowerDTO);
	        this.activeLoans = this.convertValues(source["activeLoans"], LoanDTO);
	        this.coveredAccesses = this.convertValues(source["coveredAccesses"], AccessDTO);
	        this.redundancies = this.convertValues(source["redundancies"], AccessDTO);
	        this.history = this.convertValues(source["history"], LoanDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BorrowerInput {
	    id: number;
	    name: string;
	    email: string;
	    status: string;
	    phone: string;
	
	    static createFrom(source: any = {}) {
	        return new BorrowerInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.status = source["status"];
	        this.phone = source["phone"];
	    }
	}
	export class BuildingDTO {
	    id: number;
	    name: string;
	    accessCount: number;
	
	    static createFrom(source: any = {}) {
	        return new BuildingDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.accessCount = source["accessCount"];
	    }
	}
	export class CSVImportReport {
	    created: number;
	    problems: string[];
	
	    static createFrom(source: any = {}) {
	        return new CSVImportReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.created = source["created"];
	        this.problems = source["problems"];
	    }
	}
	export class CreateLoanRequest {
	    borrowerId: number;
	    keyIds: number[];
	    plannedReturn: string;
	    loanType: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateLoanRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.borrowerId = source["borrowerId"];
	        this.keyIds = source["keyIds"];
	        this.plannedReturn = source["plannedReturn"];
	        this.loanType = source["loanType"];
	    }
	}
	export class RedundancyDTO {
	    borrowerId: number;
	    borrowerName: string;
	    keyNumbers: string[];
	    accesses: string[];
	
	    static createFrom(source: any = {}) {
	        return new RedundancyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.borrowerId = source["borrowerId"];
	        this.borrowerName = source["borrowerName"];
	        this.keyNumbers = source["keyNumbers"];
	        this.accesses = source["accesses"];
	    }
	}
	export class Dashboard {
	    keyTypes: number;
	    keyCopies: number;
	    loanedCopies: number;
	    availableCopies: number;
	    borrowers: number;
	    borrowersWithKeys: number;
	    accesses: number;
	    overdue: LoanDTO[];
	    anomalies: AnomalyDTO[];
	    redundancies: RedundancyDTO[];
	    recentLoans: LoanDTO[];
	
	    static createFrom(source: any = {}) {
	        return new Dashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keyTypes = source["keyTypes"];
	        this.keyCopies = source["keyCopies"];
	        this.loanedCopies = source["loanedCopies"];
	        this.availableCopies = source["availableCopies"];
	        this.borrowers = source["borrowers"];
	        this.borrowersWithKeys = source["borrowersWithKeys"];
	        this.accesses = source["accesses"];
	        this.overdue = this.convertValues(source["overdue"], LoanDTO);
	        this.anomalies = this.convertValues(source["anomalies"], AnomalyDTO);
	        this.redundancies = this.convertValues(source["redundancies"], RedundancyDTO);
	        this.recentLoans = this.convertValues(source["recentLoans"], LoanDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DatabaseFileInfo {
	    path: string;
	    kind: string;
	    label: string;
	    buildings: number;
	    accesses: number;
	    keys: number;
	    borrowers: number;
	    loans: number;
	    activeLoans: number;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.buildings = source["buildings"];
	        this.accesses = source["accesses"];
	        this.keys = source["keys"];
	        this.borrowers = source["borrowers"];
	        this.loans = source["loans"];
	        this.activeLoans = source["activeLoans"];
	    }
	}
	export class HistoryFilter {
	    borrowerId: number;
	    keyId: number;
	    from: string;
	    to: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.borrowerId = source["borrowerId"];
	        this.keyId = source["keyId"];
	        this.from = source["from"];
	        this.to = source["to"];
	        this.status = source["status"];
	    }
	}
	
	export class ImportResult {
	    kind: string;
	    buildings: number;
	    accesses: number;
	    keys: number;
	    borrowers: number;
	    loans: number;
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.buildings = source["buildings"];
	        this.accesses = source["accesses"];
	        this.keys = source["keys"];
	        this.borrowers = source["borrowers"];
	        this.loans = source["loans"];
	    }
	}
	export class InventoryItem {
	    keyId: number;
	    inStock: number;
	
	    static createFrom(source: any = {}) {
	        return new InventoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keyId = source["keyId"];
	        this.inStock = source["inStock"];
	    }
	}
	export class InventoryResult {
	    counted: number;
	    changed: number;
	    delta: number;
	    reportPath: string;
	    reportError: string;
	
	    static createFrom(source: any = {}) {
	        return new InventoryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.counted = source["counted"];
	        this.changed = source["changed"];
	        this.delta = source["delta"];
	        this.reportPath = source["reportPath"];
	        this.reportError = source["reportError"];
	    }
	}
	
	export class KeyDetail {
	    key: KeyDTO;
	    accesses: AccessDTO[];
	    activeLoans: LoanDTO[];
	    history: LoanDTO[];
	
	    static createFrom(source: any = {}) {
	        return new KeyDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = this.convertValues(source["key"], KeyDTO);
	        this.accesses = this.convertValues(source["accesses"], AccessDTO);
	        this.activeLoans = this.convertValues(source["activeLoans"], LoanDTO);
	        this.history = this.convertValues(source["history"], LoanDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class KeyInput {
	    id: number;
	    number: string;
	    description: string;
	    category: string;
	    storageLocation: string;
	    notes: string;
	    total: number;
	    reserve: number;
	    accessIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new KeyInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.number = source["number"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.storageLocation = source["storageLocation"];
	        this.notes = source["notes"];
	        this.total = source["total"];
	        this.reserve = source["reserve"];
	        this.accessIds = source["accessIds"];
	    }
	}
	
	export class LoanResult {
	    count: number;
	    borrowerName: string;
	    receiptPath: string;
	    receiptError: string;
	    anomalies: AnomalyDTO[];
	
	    static createFrom(source: any = {}) {
	        return new LoanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.borrowerName = source["borrowerName"];
	        this.receiptPath = source["receiptPath"];
	        this.receiptError = source["receiptError"];
	        this.anomalies = this.convertValues(source["anomalies"], AnomalyDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlanAccess {
	    access: AccessDTO;
	    keys: KeyDTO[];
	
	    static createFrom(source: any = {}) {
	        return new PlanAccess(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.access = this.convertValues(source["access"], AccessDTO);
	        this.keys = this.convertValues(source["keys"], KeyDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlanBuilding {
	    id: number;
	    name: string;
	    accesses: PlanAccess[];
	
	    static createFrom(source: any = {}) {
	        return new PlanBuilding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.accesses = this.convertValues(source["accesses"], PlanAccess);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Settings {
	    appName: string;
	    establishment: string;
	    address: string;
	    logoDataUrl: string;
	    footerNote: string;
	    receiptTitle: string;
	    engagement: string;
	    signatureBorrower: string;
	    signatureEstablishment: string;
	    signatory: string;
	    defaultLoanType: string;
	    defaultReturnDays: number;
	    agentName: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appName = source["appName"];
	        this.establishment = source["establishment"];
	        this.address = source["address"];
	        this.logoDataUrl = source["logoDataUrl"];
	        this.footerNote = source["footerNote"];
	        this.receiptTitle = source["receiptTitle"];
	        this.engagement = source["engagement"];
	        this.signatureBorrower = source["signatureBorrower"];
	        this.signatureEstablishment = source["signatureEstablishment"];
	        this.signatory = source["signatory"];
	        this.defaultLoanType = source["defaultLoanType"];
	        this.defaultReturnDays = source["defaultReturnDays"];
	        this.agentName = source["agentName"];
	    }
	}
	export class SuggestRequest {
	    borrowerId: number;
	    accessIds: number[];
	    excludedKeyIds: number[];
	    manualKeyIds: number[];
	
	    static createFrom(source: any = {}) {
	        return new SuggestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.borrowerId = source["borrowerId"];
	        this.accessIds = source["accessIds"];
	        this.excludedKeyIds = source["excludedKeyIds"];
	        this.manualKeyIds = source["manualKeyIds"];
	    }
	}
	export class SuggestedKey {
	    key: KeyDTO;
	    suggested: boolean;
	    manual: boolean;
	    covers: string[];
	
	    static createFrom(source: any = {}) {
	        return new SuggestedKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = this.convertValues(source["key"], KeyDTO);
	        this.suggested = source["suggested"];
	        this.manual = source["manual"];
	        this.covers = source["covers"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Suggestion {
	    keys: SuggestedKey[];
	    uncoverable: AccessDTO[];
	    redundancies: string[];
	
	    static createFrom(source: any = {}) {
	        return new Suggestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keys = this.convertValues(source["keys"], SuggestedKey);
	        this.uncoverable = this.convertValues(source["uncoverable"], AccessDTO);
	        this.redundancies = source["redundancies"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WhoHasWhat {
	    borrower: BorrowerDTO;
	    loans: LoanDTO[];
	    accesses: string[];
	    redundancies: string[];
	
	    static createFrom(source: any = {}) {
	        return new WhoHasWhat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.borrower = this.convertValues(source["borrower"], BorrowerDTO);
	        this.loans = this.convertValues(source["loans"], LoanDTO);
	        this.accesses = source["accesses"];
	        this.redundancies = source["redundancies"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

