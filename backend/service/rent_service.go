package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"strconv"
	"time"

	// "github.com/nescool101/rentManager/storage" // No longer directly using storage.GetPayers
	"github.com/nescool101/rentManager/model"
	"github.com/nescool101/rentManager/storage" // Added back for repository types
)

// var payers []model.Payer // This global variable is now obsolete and removed.

// GetAllPayers returns all payers from memory
// This function is obsolete as payers are now in the database.
func GetAllPayers() []model.Payer {
	log.Println("⚠️ [WARNING] GetAllPayers called - this function is obsolete.")
	return []model.Payer{} // Return empty slice
}

// LoadPayers was removed as payers are now in the database.

// NotifyAll fetches active rentals from the database and sends notifications.
// TODO: This function will require UserRepository access to fetch renter emails.
func NotifyAll(personRepo *storage.PersonRepository, rentalRepo *storage.RentalRepository, propertyRepo *storage.PropertyRepository, userRepo *storage.UserRepository, pricingRepo *storage.PricingRepository) {
	ctx := context.Background()
	loc, _ := time.LoadLocation("America/New_York") // Consider making timezone configurable
	today := time.Now().In(loc)

	log.Println("ℹ️ [INFO] NotifyAll: Starting notification process...")

	activeRentals, err := rentalRepo.GetActiveRentals(ctx) // Assuming GetActiveRentals doesn't need a specific user context for a system-wide job
	if err != nil {
		log.Printf("❌ [ERROR] NotifyAll: Failed to fetch active rentals: %v", err)
		return
	}

	if len(activeRentals) == 0 {
		log.Println("ℹ️ [INFO] NotifyAll: No active rentals found to process.")
		return
	}

	log.Printf("ℹ️ [INFO] NotifyAll: Found %d active rentals to process.", len(activeRentals))

	for _, rental := range activeRentals {
		renter, err := personRepo.GetByID(ctx, rental.RenterID)
		if err != nil {
			log.Printf("⚠️ [WARNING] NotifyAll: Failed to fetch renter (person_id: %s) for rental_id %s: %v. Skipping rental.", rental.RenterID, rental.ID, err)
			continue
		}
		if renter == nil || renter.FullName == "" {
			log.Printf("⚠️ [WARNING] NotifyAll: Renter (person_id: %s) not found or has no name for rental_id %s. Skipping rental.", rental.RenterID, rental.ID)
			continue
		}

		// Fetch user email via UserRepository
		renterUser, userErr := userRepo.GetByPersonID(ctx, renter.ID)
		if userErr != nil {
			log.Printf("⚠️ [WARNING] NotifyAll: Error fetching user record for renter (person_id: %s): %v. Skipping rental.", renter.ID, userErr)
			continue
		}
		if renterUser == nil || renterUser.Email == "" {
			log.Printf("⚠️ [WARNING] NotifyAll: User record not found or email is missing for renter (person_id: %s). Skipping rental.", renter.ID)
			continue
		}
		renterEmail := renterUser.Email

		log.Printf("ℹ️ [INFO] NotifyAll: Processing for renter %s (PersonID: %s, Email: %s)", renter.FullName, renter.ID, renterEmail)

		property, err := propertyRepo.GetByID(ctx, rental.PropertyID)
		if err != nil {
			log.Printf("⚠️ [WARNING] NotifyAll: Failed to fetch property (property_id: %s) for rental_id %s: %v. Skipping rental.", rental.PropertyID, rental.ID, err)
			continue
		}
		if property == nil {
			log.Printf("⚠️ [WARNING] NotifyAll: Property (property_id: %s) not found for rental_id %s. Skipping rental.", rental.PropertyID, rental.ID)
			continue
		}

		// Fetch pricing information for the rental
		pricing, pricingErr := pricingRepo.GetByRentalID(ctx, rental.ID)
		if pricingErr != nil {
			log.Printf("⚠️ [WARNING] NotifyAll: Error fetching pricing for rental_id %s: %v. Skipping rental.", rental.ID, pricingErr)
			continue
		}
		if pricing == nil {
			log.Printf("⚠️ [WARNING] NotifyAll: Pricing information not found for rental_id %s. Skipping rental.", rental.ID)
			continue
		}

		senderName := "La Administración"
		if len(property.ManagerIDs) > 0 {
			firstManager, mErr := personRepo.GetByID(ctx, property.ManagerIDs[0])
			if mErr == nil && firstManager != nil {
				senderName = firstManager.FullName
			} else {
				log.Printf("⚠️ [WARNING] NotifyAll: Could not fetch manager details for property %s. Using default sender.", property.ID)
			}
		}

		rentalStartDate := rental.StartDate.Time() // Use .Time() method of FlexibleTime

		rentalDay := rentalStartDate.Day()
		rentalMonth := rentalStartDate.Month()
		rentalYear := rentalStartDate.Year()

		log.Printf("Processing rental %s for renter %s, property %s. Start Date: %s (Day: %d, Month: %s, Year: %d)",
			rental.ID, renterEmail, property.Address, rental.StartDate.Time().Format(time.RFC3339), rentalDay, rentalMonth.String(), rentalYear)

		// Call refactored reminder functions
		sendSameMonthReminderEmail(today, pricing.DueDay, &rental, renter, property, senderName, renterEmail, pricing)
		sendSameYearReminderEmail(today, rentalDay, rentalMonth, rentalYear, renter, property, senderName, renterEmail)

		// _ = today             // Suppress unused error for now
		// _ = senderName        // Suppress unused error for now
	}
}

// Send one-year rental anniversary reminder
// TODO: Refactor this function to accept model.Rental, model.Person (renter), model.Property, senderName string
func sendSameYearReminderEmail(today time.Time, rentalDay int, rentalMonth time.Month, rentalYear int, renter *model.Person, property *model.Property, senderName string, renterEmail string) {
	if today.Day() == rentalDay && today.Month() == rentalMonth && today.Year() != rentalYear {
		log.Printf("📩 [1-YEAR ANNIVERSARY] Preparing for: Renter %s (%s), Property %s",
			renter.FullName, renterEmail, property.Address)

		subject := "🏡 Aniversario de Arrendamiento"
		body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
		    <meta charset="UTF-8">
		    <title>Aniversario de Arrendamiento</title>
		    <style>
		        body { font-family: Arial, sans-serif; }
		        .container { padding: 20px; }
		        .highlight { font-weight: bold; color: #007BFF; }
		    </style>
		</head>
		<body>
		    <div class="container">
		        <h2>🏡 ¡Feliz Aniversario de Arrendamiento, %s!</h2>
		        <p>Hoy se cumple un año desde que inició su contrato de arrendamiento para la propiedad en:</p>
		        <p class="highlight">%s</p>
		        <p>Le agradecemos su confianza y esperamos que su experiencia haya sido excelente.</p>
		        <p>¿Desea renovar su contrato de arrendamiento?</p>
		        <p>Por favor, comuníquese con nosotros para discutir las opciones de renovación.</p>
		        <hr>
		        <p>Atentamente,</p>
		        <p><strong>%s</strong></p>
		    </div>
		</body>
		</html>
		`, renter.FullName, property.Address, senderName)

		// Send email to Tenant (Renter)
		err := SendSimpleEmail(renterEmail, subject, body)
		if err != nil {
			log.Printf("❌ [FAILED] 1-Year Anniversary Email NOT sent to Renter: %s (%s) - Error: %v",
				renter.FullName, renterEmail, err)
		} else {
			log.Printf("✅ [SENT] 1-Year Anniversary Email sent to Renter: %s (%s)",
				renter.FullName, renterEmail)
		}
	}
}

// Send one-month rental reminder
// TODO: Refactor this function to accept model.Rental, model.Person (renter), model.Property, senderName string
func sendSameMonthReminderEmail(today time.Time, dueDay int, rental *model.Rental, renter *model.Person, property *model.Property, senderName string, renterEmail string, pricing *model.Pricing) {
	if today.Day() == dueDay { // Use dueDay from pricing
		log.Printf("📩 [MONTHLY RENT REMINDER] Preparing for: Renter %s (%s), Property %s", renter.FullName, renterEmail, property.Address)

		// Construct Payer-like object for template, or adapt template directly
		// For now, let's adapt key fields for sendEmail which expects model.Payer
		payerForEmail := model.Payer{
			Name:            renter.FullName,
			RentalEmail:     renterEmail,
			PropertyAddress: property.Address,
			MonthlyRent:     int(pricing.MonthlyRent), // Use MonthlyRent from pricing
			// DueDate: rental.StartDate.Time().Format("January 2, 2006"), // TODO: Construct actual due date for current month using pricing.DueDay
			RenterName:   senderName,              // This is the email sender, effectively
			RentalDate:   rental.StartDate.Time(), // Pass rental start date
			NIT:          renter.NIT,              // Pass renter's NIT
			PropertyType: property.Type,           // Pass property type
			RentalStart:  rental.StartDate.Time(),
			RentalEnd:    rental.EndDate.Time(),
			PaymentTerms: rental.PaymentTerms,
			UnpaidMonths: rental.UnpaidMonths, // This comes from Rental model
		}

		err := sendEmail(renterEmail, payerForEmail) // sendEmail still expects a model.Payer
		if err != nil {
			log.Printf("❌ [FAILED] Monthly Rent Reminder NOT sent to %s (%s) - Error: %v", renter.FullName, renterEmail, err)
			return
		}
		log.Printf("✅ [SENT] Monthly Rent Reminder sent to: %s (%s) for property %s", renter.FullName, renterEmail, property.Address)
	} else {
		// This log might be too verbose if NotifyAll runs daily. Consider removing or reducing its frequency.
		// log.Printf("Skipping monthly reminder for %s (%s) - Day %d != %d", renter.FullName, renterEmail, today.Day(), rentalDay)
	}
}

// EmailTemplate represents the structure of the email data
type EmailTemplate struct {
	EmisorNombre         string
	EmisorNIT            string
	EmisorDireccion      string
	EmisorTelefono       string
	EmisorEmail          string
	NumeroCuenta         int
	FechaEmision         string
	ArrendatarioNombre   string
	ArrendatarioNIT      string
	InmuebleDireccion    string
	TipoInmueble         string
	FechaInicio          string
	FechaFinal           string
	ValorMensual         string
	Subtotal             string
	TotalPagar           string
	CondicionesPago      string
	Banco                string
	TipoCuenta           string
	NumeroCuentaBancaria string
	TitularCuenta        string
	Observaciones        string
	ArrendadorNombre     string
	UnpaidMonths         int
	TotalDue             string
}

// Email template in HTML format
const emailTemplateHTML = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Cuenta de Cobro</title>
</head>
<body>
    <hr>
    <h3>CUENTA DE COBRO ARRENDAMIENTO N° {{.NumeroCuenta}}</h3>
    <p>Fecha: {{.FechaEmision}}</p>
    <h4>Informacion de arrendatario:</h4>
    <p>Nombre del Arrendatario: {{.ArrendatarioNombre}}</p>
    <p>NIT/Cédula del Arrendatario: {{.ArrendatarioNIT}}</p>
    <p>Dirección del Inmueble Arrendado: {{.InmuebleDireccion}}</p>
    <hr>
    <h3>Descripción del Arrendamiento:</h3>
    <table border="1">
        <tr>
            <th>Tipo de Inmueble</th>
            <th>Fecha Inicio</th>
            <th>Fecha Final</th>
            <th>Valor Mensual</th>
            <th>Subtotal</th>
        </tr>
        <tr>
            <td>{{.TipoInmueble}}</td>
            <td>{{.FechaInicio}}</td>
            <td>{{.FechaFinal}}</td>
            <td>{{.ValorMensual}}</td>
            <td>{{.Subtotal}}</td>
        </tr>
    </table>
    <h3>Total a Pagar: {{.TotalPagar}}</h3>
    {{if gt .UnpaidMonths 0}}
        <div class="highlight">
            <h3 class="warning">⚠️ Pagos Atrasados</h3>
            <p>El arrendatario tiene <strong>{{.UnpaidMonths}} meses</strong> sin pagar.</p>
            <p>Monto total adeudado: <strong>{{.TotalDue}}</strong></p>
            <p>Por favor, realice el pago lo antes posible para evitar sanciones.</p>
        </div>
        <hr>
    {{end}}

    <hr>
    <h4>Condiciones de Pago:</h4>
    <p>{{.CondicionesPago}}</p>
    <h4>Datos Bancarios para Transferencias:</h4>
    <p>Banco: {{.Banco}}</p>
    <p>Tipo de Cuenta: {{.TipoCuenta}}</p>
    <p>Número de Cuenta: {{.NumeroCuentaBancaria}}</p>
    <p>Titular de la Cuenta: {{.TitularCuenta}}</p>
    <h4>Observaciones Adicionales:</h4>
    <p>{{.Observaciones}}</p>
    <hr>
    <p>Atentamente,</p>
    <p>{{.ArrendadorNombre}}</p>
</body>
</html>
`

func sendEmail(to string, payer model.Payer) error {
	// Convert MonthlyRent to an integer (removing "USD" or currency text)
	totalDue := 0
	if payer.UnpaidMonths > 0 {
		totalDue = payer.MonthlyRent * payer.UnpaidMonths
	}

	data := EmailTemplate{
		EmisorNombre:         "Mi Empresa S.A.",
		EmisorNIT:            "123456789",
		EmisorDireccion:      "Calle 123, Ciudad",
		EmisorTelefono:       "555-1234",
		EmisorEmail:          "empresa@example.com",
		NumeroCuenta:         rentalDateToInt(payer.RentalDate),
		FechaEmision:         payer.RentalDate.Format("02/01/2006"),
		ArrendatarioNombre:   payer.Name,
		ArrendatarioNIT:      payer.NIT,
		InmuebleDireccion:    payer.PropertyAddress,
		TipoInmueble:         payer.PropertyType,
		FechaInicio:          payer.RentalStart.Format("02/01/2006"),
		FechaFinal:           payer.RentalEnd.Format("02/01/2006"),
		ValorMensual:         strconv.Itoa(payer.MonthlyRent),
		Subtotal:             strconv.Itoa(payer.MonthlyRent),
		TotalPagar:           strconv.Itoa(payer.MonthlyRent),
		CondicionesPago:      "Pago antes del 5 de cada mes",
		Banco:                payer.BankName,
		TipoCuenta:           payer.AccountType,
		NumeroCuentaBancaria: payer.BankAccountNumber,
		TitularCuenta:        payer.AccountHolder,
		Observaciones:        payer.AdditionalNotes,
		ArrendadorNombre:     payer.RenterName,
		UnpaidMonths:         payer.UnpaidMonths,
		TotalDue:             strconv.Itoa(totalDue) + " COP",
	}

	// Parse and execute the HTML template
	tmpl, err := template.New("email").Parse(emailTemplateHTML)
	if err != nil {
		log.Println("Error parsing template:", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Println("Error executing template:", err)
		return err
	}

	// Use our new ProtonMail implementation instead of Resend
	err = SendProtonMailEmail(to, "Cuenta de Cobro Arrendamiento", body.String())
	if err != nil {
		log.Printf("❌ [EMAIL NOT SENT] %s (%s) - Error: %v", payer.Name, to, err)
		return err
	}

	log.Printf("✅ [EMAIL SENT] %s (%s)", payer.Name, to)
	return nil
}

func rentalDateToInt(date time.Time) int {
	return date.Year()*10000 + int(date.Month())*100 + date.Day()
}

// SendAnnualRenewalReminders sends reminders to tenants whose contracts are ending in approximately one month.
func SendAnnualRenewalReminders(ctx context.Context, personRepo *storage.PersonRepository, rentalRepo *storage.RentalRepository, propertyRepo *storage.PropertyRepository, userRepo *storage.UserRepository, optionalMessage string) (int, error) {
	loc, _ := time.LoadLocation("America/New_York")      // Consider making timezone configurable
	today := time.Now().In(loc).Truncate(24 * time.Hour) // Truncate to just the date part
	targetEndDateLowerBound := today.AddDate(0, 1, -2)   // Approx 1 month from today, with a small window (e.g., 28 days)
	targetEndDateUpperBound := today.AddDate(0, 1, 2)    // Approx 1 month from today, with a small window (e.g., 32 days)

	log.Printf("ℹ️ [ANNUAL REMINDER] Starting process. Target EndDate window: %s to %s", targetEndDateLowerBound.Format("2006-01-02"), targetEndDateUpperBound.Format("2006-01-02"))

	activeRentals, err := rentalRepo.GetActiveRentals(ctx)
	if err != nil {
		log.Printf("❌ [ERROR] SendAnnualRenewalReminders: Failed to fetch active rentals: %v", err)
		return 0, fmt.Errorf("failed to fetch active rentals: %w", err)
	}

	if len(activeRentals) == 0 {
		log.Println("ℹ️ [INFO] SendAnnualRenewalReminders: No active rentals found.")
		return 0, nil
	}

	emailsSent := 0
	for _, rental := range activeRentals {
		rentalEndDate := rental.EndDate.Time().In(loc).Truncate(24 * time.Hour)

		// Check if the rental end date is within our target window (approx. 1 month from now)
		if (rentalEndDate.After(targetEndDateLowerBound) || rentalEndDate.Equal(targetEndDateLowerBound)) &&
			(rentalEndDate.Before(targetEndDateUpperBound) || rentalEndDate.Equal(targetEndDateUpperBound)) {

			log.Printf("Processing rental %s ending on %s for annual renewal reminder.", rental.ID, rentalEndDate.Format("2006-01-02"))

			renter, pErr := personRepo.GetByID(ctx, rental.RenterID)
			if pErr != nil || renter == nil {
				log.Printf("⚠️ [WARNING] SendAnnualRenewalReminders: Failed to fetch renter for rental_id %s: %v. Skipping.", rental.ID, pErr)
				continue
			}

			renterUser, uErr := userRepo.GetByPersonID(ctx, renter.ID)
			if uErr != nil || renterUser == nil || renterUser.Email == "" {
				log.Printf("⚠️ [WARNING] SendAnnualRenewalReminders: User/email not found for renter %s (PersonID: %s). Skipping.", renter.FullName, renter.ID)
				continue
			}

			property, propErr := propertyRepo.GetByID(ctx, rental.PropertyID)
			if propErr != nil || property == nil {
				log.Printf("⚠️ [WARNING] SendAnnualRenewalReminders: Property not found for rental_id %s. Skipping.", rental.ID)
				continue
			}

			senderName := "La Administración"
			if len(property.ManagerIDs) > 0 {
				firstManager, mErr := personRepo.GetByID(ctx, property.ManagerIDs[0])
				if mErr == nil && firstManager != nil {
					senderName = firstManager.FullName
				}
			}

			subject := fmt.Sprintf("Reminder: Your Lease for %s is Ending Soon", property.Address)
			bodyText := fmt.Sprintf(
				`<!DOCTYPE html><html><head><title>%s</title></head><body>
				<p>Dear %s,</p>
				<p>This is a friendly reminder that your lease agreement for the property at <strong>%s</strong> is scheduled to end on <strong>%s</strong>.</p>
				<p>We value you as a tenant and would like to invite you to discuss renewal options. Please contact us at your earliest convenience to explore continuing your stay.</p>`,
				subject, renter.FullName, property.Address, rental.EndDate.Time().Format("January 2, 2006"))

			if optionalMessage != "" {
				bodyText += fmt.Sprintf("<p><strong>Additional message from administration:</strong><br>%s</p>", optionalMessage)
			}
			bodyText += fmt.Sprintf("<p>Sincerely,</p><p>%s</p></body></html>", senderName)

			if err := SendSimpleEmail(renterUser.Email, subject, bodyText); err == nil {
				log.Printf("✅ [ANNUAL REMINDER SENT] To: %s for property %s", renterUser.Email, property.Address)
				emailsSent++
			} else {
				log.Printf("❌ [ANNUAL REMINDER FAILED] To: %s for property %s - Error: %v", renterUser.Email, property.Address, err)
			}
		} // end if rental end date in window
	} // end for rental

	log.Printf("ℹ️ [ANNUAL REMINDER] Process finished. %d emails sent.", emailsSent)
	return emailsSent, nil
}
