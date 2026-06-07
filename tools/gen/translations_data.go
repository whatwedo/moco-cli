package main

// nouns holds the German resource nouns per OpenAPI tag, used to build the
// formulaic German descriptions for the standard CRUD commands.
//
// The terms are taken from the "<sub>Synonyme: ...</sub>" hints in the tag
// descriptions of api/openapi-v1.yaml, which carry MOCO's official German
// terminology.
var nouns = map[string]Noun{
	"AccountCatalogServices":     {"Leistungskatalog", "Leistungskataloge"},
	"AccountCustomProperties":    {"Eigenes Feld", "Eigene Felder"},
	"AccountExpenseTemplates":    {"Standard-Zusatzleistung", "Standard-Zusatzleistungen"},
	"AccountRates":               {"Stundensatz", "Stundensätze"},
	"AccountTaskTemplates":       {"Standardleistung", "Standardleistungen"},
	"AccountWebHooks":            {"Automatisierung", "Automatisierungen"},
	"Activities":                 {"Zeiteintrag", "Zeiteinträge"},
	"Comments":                   {"Kommentar", "Kommentare"},
	"Companies":                  {"Firma", "Firmen"},
	"Contacts":                   {"Kontakt", "Kontakte"},
	"Contracts":                  {"Projektzuweisung", "Projektzuweisungen"},
	"DealCategories":             {"Akquise-Stufe", "Akquise-Stufen"},
	"Deals":                      {"Lead", "Leads"},
	"Expenses":                   {"Spese", "Spesen"},
	"InvoiceAttachments":         {"Rechnungsanhang", "Rechnungsanhänge"},
	"InvoiceBookkeepingExports":  {"Buchhaltungsexport", "Buchhaltungsexporte"},
	"InvoicePayments":            {"Zahlung", "Zahlungen"},
	"InvoiceReminders":           {"Zahlungserinnerung", "Zahlungserinnerungen"},
	"Invoices":                   {"Rechnung", "Rechnungen"},
	"LetterPapers":               {"Briefpapier", "Briefpapiere"},
	"OfferAttachments":           {"Angebotsanhang", "Angebotsanhänge"},
	"OfferCustomerApproval":      {"Angebotsbestätigung", "Angebotsbestätigungen"},
	"Offers":                     {"Angebot", "Angebote"},
	"PlanningEntries":            {"Planungseintrag", "Planungseinträge"},
	"Profile":                    {"Profil", "Profile"},
	"ProjectGroups":              {"Projektgruppe", "Projektgruppen"},
	"ProjectPaymentSchedules":    {"Abrechnungsplan", "Abrechnungspläne"},
	"Projects":                   {"Projekt", "Projekte"},
	"PurchaseBookkeepingExports": {"Buchhaltungsexport", "Buchhaltungsexporte"},
	"PurchaseBudgets":            {"Ausgabenbudget", "Ausgabenbudgets"},
	"PurchaseCategories":         {"Aufwandskategorie", "Aufwandskategorien"},
	"PurchaseDrafts":             {"Ausgabenentwurf", "Ausgabenentwürfe"},
	"PurchaseOrders":             {"Bestellung", "Bestellungen"},
	"PurchasePayments":           {"Zahlung", "Zahlungen"},
	"Purchases":                  {"Ausgabe", "Ausgaben"},
	"Receipts":                   {"Beleg", "Belege"},
	"RecurringExpenses":          {"Wiederkehrende Ausgabe", "Wiederkehrende Ausgaben"},
	"Reports":                    {"Bericht", "Berichte"},
	"Schedules":                  {"Abwesenheit", "Abwesenheiten"},
	"Session":                    {"Sitzung", "Sitzungen"},
	"Taggings":                   {"Label", "Labels"},
	"Tags":                       {"Label", "Labels"},
	"Tasks":                      {"Leistung", "Leistungen"},
	"Units":                      {"Team", "Teams"},
	"UserEmployments":            {"Wochenmodell", "Wochenmodelle"},
	"UserHolidays":               {"Urlaubsanspruch", "Urlaubsansprüche"},
	"UserPresences":              {"Arbeitszeit", "Arbeitszeiten"},
	"UserRoles":                  {"Benutzerrolle", "Benutzerrollen"},
	"Users":                      {"Person", "Personen"},
	"UserWorkTimeAdjustments":    {"Korrektur", "Korrekturen"},
	"VatCodePurchases":           {"Steuercode", "Steuercodes"},
	"VatCodeSales":               {"Steuercode", "Steuercodes"},
}

// actions holds fixed German descriptions for non-CRUD commands and for CRUD
// commands on sub-resources, keyed by "<group>/<name>". The resource terms
// follow the same MOCO terminology as the nouns above.
var actions = map[string]string{
	"account-catalog-services/items-create": "Leistungskatalog-Position erstellen",
	"account-catalog-services/items-get":    "Leistungskatalog-Position abrufen",
	"account-catalog-services/items-update": "Leistungskatalog-Position aktualisieren",
	"account-catalog-services/items-patch":  "Leistungskatalog-Position teilweise aktualisieren",
	"account-catalog-services/items-delete": "Leistungskatalog-Position löschen",

	"account-rates/fixed-costs-list":             "Fixkosten auflisten",
	"account-rates/hourly-rates-list":            "Stundensätze auflisten",
	"account-rates/internal-hourly-rates-get":    "Interne Stundensätze abrufen",
	"account-rates/internal-hourly-rates-update": "Interne Stundensätze aktualisieren",
	"account-rates/internal-hourly-rates-patch":  "Interne Stundensätze teilweise aktualisieren",

	"account-web-hooks/disable": "Automatisierung deaktivieren",
	"account-web-hooks/enable":  "Automatisierung aktivieren",

	"activities/bulk-create":             "Zeiteinträge als Stapel erstellen",
	"activities/disregard":               "Zeiteinträge von der Verrechnung ausnehmen",
	"activities/update-billable-seconds": "Verrechenbare Sekunden des Zeiteintrags aktualisieren",
	"activities/start-timer":             "Zeiterfassungs-Timer starten",
	"activities/stop-timer":              "Zeiterfassungs-Timer stoppen",

	"comments/bulk-create": "Kommentare als Stapel erstellen",

	"companies/archive":   "Firma archivieren",
	"companies/unarchive": "Firma aus dem Archiv holen",

	"invoice-reminders/send-email": "Zahlungserinnerung per E-Mail senden",

	"invoices/locked-list":   "Gesperrte Rechnungen auflisten",
	"invoices/pdf":           "Rechnungs-PDF herunterladen",
	"invoices/expenses-list": "Rechnungs-Spesen auflisten",
	"invoices/send-email":    "Rechnung per E-Mail senden",
	"invoices/timesheet-get": "Leistungsnachweis der Rechnung abrufen",
	"invoices/timesheet-pdf": "Leistungsnachweis-PDF der Rechnung herunterladen",
	"invoices/update-status": "Rechnungsstatus aktualisieren",

	"invoice-payments/bulk-create": "Zahlungen als Stapel erstellen",

	"offers/pdf":           "Angebots-PDF herunterladen",
	"offers/assign":        "Angebot einer Firma/einem Projekt/Deal zuordnen",
	"offers/send-email":    "Angebot per E-Mail senden",
	"offers/update-status": "Angebotsstatus aktualisieren",

	"offer-customer-approval/activate":   "Kundenfreigabe aktivieren",
	"offer-customer-approval/deactivate": "Kundenfreigabe deaktivieren",

	"projects/assigned-list":          "Mir zugewiesene Projekte auflisten",
	"projects/archive":                "Projekt archivieren",
	"projects/assign-project-group":   "Projekt einer Projektgruppe zuordnen",
	"projects/disable-share":          "Projektbericht-Freigabe deaktivieren",
	"projects/report-get":             "Projektbericht abrufen",
	"projects/share":                  "Projektbericht-Freigabe aktivieren",
	"projects/unarchive":              "Projekt aus dem Archiv holen",
	"projects/unassign-project-group": "Projekt aus Projektgruppe entfernen",

	"expenses/list-all":    "Spesen über alle Projekte auflisten",
	"expenses/bulk-create": "Projekt-Spesen als Stapel erstellen",
	"expenses/disregard":   "Projekt-Spesen von der Verrechnung ausnehmen",

	"project-payment-schedules/list-all": "Abrechnungspläne aller Projekte auflisten",

	"recurring-expenses/recur":    "Wiederkehrende Ausgabe erneut anlegen",
	"recurring-expenses/list-all": "Wiederkehrende Ausgaben auflisten",

	"tasks/destroy-all": "Alle Projekt-Leistungen löschen",

	"purchase-drafts/pdf": "Ausgabenentwurf-PDF herunterladen",

	"purchase-payments/bulk-create": "Zahlungen als Stapel erstellen",

	"purchases/assign-to-project": "Ausgabenposition einem Projekt zuordnen",
	"purchases/store-document":    "Ausgabenbeleg ablegen",
	"purchases/update-status":     "Ausgabenstatus aktualisieren",

	"reports/absences-get":           "Abwesenheitsbericht abrufen",
	"reports/cashflow-get":           "Cashflow-Bericht abrufen",
	"reports/finance-get":            "Finanzbericht abrufen",
	"reports/planned-vs-tracked-get": "Bericht „Geplant vs. Erfasst“ abrufen",
	"reports/utilization-get":        "Auslastungsbericht abrufen",

	"user-presences/touch":         "Arbeitszeit erfassen (touch)",
	"users/performance-report-get": "Leistungsbericht der Person abrufen",
}

// nameOverrides renames individual commands whose path would otherwise yield an
// unclear or colliding name. Key: "<METHOD> <path>".
var nameOverrides = map[string]string{
	// Global aggregation endpoints (without {project_id}) vs. their
	// project-scoped siblings.
	"GET /projects/expenses":          "list-all",
	"GET /projects/payment_schedules": "list-all",
	"GET /recurring_expenses":         "list-all",

	// Action paths whose trailing literal already implies a verb; avoid the
	// doubled "...-update"/"...-patch"/"...-delete" suffix.
	"PUT /invoices/{id}/update_status":                "update-status",
	"PUT /offers/{id}/update_status":                  "update-status",
	"PATCH /purchases/{id}/update_status":             "update-status",
	"PATCH /activities/{id}/billable_seconds":         "update-billable-seconds",
	"DELETE /projects/{project_id}/tasks/destroy_all": "destroy-all",
}
