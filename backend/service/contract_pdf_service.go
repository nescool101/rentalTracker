package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/nescool101/rentManager/model"
)

// ContractPDF represents the data needed to generate a rental contract PDF
type ContractPDF struct {
	Renter         *model.Person
	Owner          *model.Person
	Property       *model.Property
	Pricing        *model.Pricing
	CoSigner       *model.Person // Deudor solidario
	Witness        *model.Person // Testigo
	StartDate      time.Time
	EndDate        time.Time
	AdditionalInfo string
	CreationDate   time.Time
	DepositText    string // Text describing deposit conditions
}

// GenerateContractPDF creates a rental contract PDF using the complete Colombian template
func GenerateContractPDF(data ContractPDF) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set up basic formatting
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Extract data with defaults
	currentDate := FormatSpanishDate(data.CreationDate)
	propertyAddress := getPropertyAddress(data.Property)
	garageNumber := getGarageNumber(data.Property)
	buildingName := getBuildingName(data.Property)

	arrendadorName := "MARIA VICTORIA JIMENEZ DE ROSAS"
	arrendadorCC := "41.350.115"
	if data.Owner != nil && data.Owner.FullName != "" {
		arrendadorName = strings.ToUpper(data.Owner.FullName)
		arrendadorCC = data.Owner.NIT
	}

	arrendatarioName := "SEBASTIÁN MOTAVITA MEDELLÍN"
	arrendatarioCC := "1.026.281.306"
	if data.Renter != nil && data.Renter.FullName != "" {
		arrendatarioName = strings.ToUpper(data.Renter.FullName)
		arrendatarioCC = data.Renter.NIT
	}

	testigoName := "NA"
	testigoCC := "NA"
	if data.Witness != nil && data.Witness.FullName != "" {
		testigoName = strings.ToUpper(data.Witness.FullName)
		testigoCC = data.Witness.NIT
	}

	codeudorName := "NÉSTOR FERNANDO ÁLVAREZ"
	codeudorCC := "1.015.398.879"
	if data.CoSigner != nil && data.CoSigner.FullName != "" {
		codeudorName = strings.ToUpper(data.CoSigner.FullName)
		codeudorCC = data.CoSigner.NIT
	}

	canonMensual := "$1,600,000.00"
	canonIncluido := "INCLUÍDA LA ADMINISTRACIÓN"
	if data.Pricing != nil && data.Pricing.MonthlyRent > 0 {
		canonMensual = FormatMoney(data.Pricing.MonthlyRent)
	}

	fechaIniciacion := "Junio 6 de 2022"
	fechaTerminacion := "Diciembre 5 de 2022"
	if !data.StartDate.IsZero() {
		fechaIniciacion = FormatSpanishDate(data.StartDate)
	}
	if !data.EndDate.IsZero() {
		fechaTerminacion = FormatSpanishDate(data.EndDate)
	}

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.MultiCell(0, 8, fixSpanishChars("CONTRATO DE ARRENDAMIENTO DE INMUEBLE PARA VIVIENDA URBANA"), "", "C", false)
	pdf.Ln(2)
	pdf.SetFont("Arial", "B", 12)
	pdf.MultiCell(0, 6, fixSpanishChars(propertyAddress), "", "C", false)
	pdf.Ln(10)

	// Header information
	pdf.SetFont("Arial", "B", 10)
	addInfoLine(pdf, "LUGAR Y FECHA DEL CONTRATO:", "Bogotá, D. C., "+currentDate)
	addInfoLine(pdf, "DIRECCION DEL INMUEBLE:", propertyAddress+",")
	addInfoLine(pdf, "", "Garaje # "+garageNumber+", Edificio "+buildingName)
	addInfoLine(pdf, "ARRENDADOR:", arrendadorName+", CC "+arrendadorCC)
	addInfoLine(pdf, "ARRENDATARIO:", arrendatarioName+", CC "+arrendatarioCC)
	addInfoLine(pdf, "TESTIGO:", testigoName+", CC "+testigoCC)
	addInfoLine(pdf, "CODEUDOR:", codeudorName+", CC "+codeudorCC)
	addInfoLine(pdf, "CANON MENSUAL:", canonMensual+" "+canonIncluido)
	addInfoLine(pdf, "FECHA INICIACION:", fechaIniciacion)
	addInfoLine(pdf, "FECHA TERMINACION:", fechaTerminacion)

	pdf.Ln(10)

	// Main content title
	pdf.SetFont("Arial", "B", 12)
	pdf.MultiCell(0, 8, fixSpanishChars("CONDICIONES GENERALES"), "", "C", false)
	pdf.Ln(5)

	// All clauses from the Colombian template
	addClause(pdf, "PRIMERA: OBJETO DEL CONTRATO:",
		"Mediante el presente contrato el ARRENDADOR concede al ARRENDATARIO el goce de los inmuebles que adelante se identifican por su dirección y linderos, de acuerdo con el inventario que las partes firman por separado, el cual forma parte integral de este mismo contrato de arrendamiento.")

	aptNumText := "DOSCIENTOS UNO (201)"
	garageNumText := "VEINTIDOS (22)"
	if data.Property != nil && data.Property.AptNumber != "" {
		aptNumText = data.Property.AptNumber
		garageNumText = data.Property.AptNumber
	}

	secondClauseText := fmt.Sprintf("APARTAMENTO %s y Garaje número %s ubicados en la %s Edificio %s, de la ciudad de Bogotá, D. C., cuyos linderos se encuentran plasmados en la Escritura Pública número 158 del 26 de Enero de 2015, otorgada en la Notaria (47) cuarenta y siete de Bogotá, según ANOTACION 008 Fecha: 03-02-2015 Radicación: 2015-7187 del Certificado de Tradición. A estos inmuebles les corresponden los folios de matrículas inmobiliarias # 50N-20677668 y # 50N-20677602, respectivamente, de la oficina de Registro de Instrumentos Públicos de Bogotá, D.C.", aptNumText, garageNumText, propertyAddress, buildingName)

	addClause(pdf, "SEGUNDA: DIRECCIÓN DE LOS INMUEBLES:", secondClauseText)

	addClause(pdf, "TERCERA: DESTINACIÓN:",
		"El ARRENDATARIO se compromete a destinar este inmueble exclusivamente para vivienda.")

	canonInWords := "UN MILLON SEISCIENTOS MIL PESOS MONEDA LEGAL"
	if data.Pricing != nil && data.Pricing.MonthlyRent > 0 {
		canonInWords = AmountInWords(data.Pricing.MonthlyRent) + " PESOS MONEDA LEGAL"
	}

	fourthClauseText := fmt.Sprintf("El canon mensual de arrendamiento, incluida la cuota de administración de la copropiedad será la suma de %s (%s) pagaderos al ARRENDADOR o a su orden en la siguiente forma: a) La suma de $4.800.000.00 (valor correspondiente a tres meses de arriendo de Junio, Julio y Agosto) a la firma del presente contrato y entrega del inmueble. b) La suma de $4.800.000 dentro de los primeros cinco (5) días del mes de septiembre 2022, correspondientes a los últimos tres meses (septiembre, octubre y noviembre) de arrendamiento. PARAGRAFO: El ARRENDATARIO pagará las sumas arriba indicadas al ARRENDADOR mediante consignación o transferencia a la CUENTA CORRIENTE No. 014366744 del Banco ITAU a nombre del ARRENDADOR, María Victoria Jiménez de Rosas identificada con la cédula No. 41350115.", canonInWords, canonMensual)

	addClause(pdf, "CUARTA PRECIO DEL ARRENDAMIENTO:", fourthClauseText)

	durationText := "SEIS (6) MESES"
	startDateText := "SEIS (6) DE JUNIO DEL AÑO 2022"
	endDateText := "cinco (5) de Diciembre de 2022"
	if !data.StartDate.IsZero() && !data.EndDate.IsZero() {
		duration := int(data.EndDate.Sub(data.StartDate).Hours() / 24 / 30.44) // More accurate month calculation
		durationText = fmt.Sprintf("%s (%s) MESES", NumberToWords(duration), strings.ToUpper(NumberToWords(duration)))
		startDateText = FormatSpanishDateWithDay(data.StartDate)
		endDateText = FormatSpanishDate(data.EndDate)
	}

	fifthClauseText := fmt.Sprintf("La vigencia del presente contrato será de %s, a partir del %s y hasta la fecha de entrega de los inmuebles que deberá ser el día %s, salvo lo acordado en la Cláusula Octava: PRORROGAS.", durationText, startDateText, endDateText)

	addClause(pdf, "QUINTA: VIGENCIA DEL CONTRATO:", fifthClauseText)

	addClause(pdf, "SEXTA. CUOTAS DE ADMINISTRACIÓN:",
		fmt.Sprintf("La cuota ordinaria mensual de administración será cancelada por el ARRENDADOR directamente a la Copropiedad dentro de los plazos fijados por la Copropiedad en la respectiva factura. PARAGRAFO UNO: EL ARRENDATARIO se compromete a cumplir y respetar cabalmente todas y cada una de las normas establecidas por el Reglamento de Propiedad Horizontal y del Manual de Convivencia del Edificio %s.", buildingName))

	addClause(pdf, "SEPTIMA- INCREMENTOS DEL PRECIO:",
		"Vencidos los seis (6) meses de vigencia de este contrato y así sucesivamente cada seis (6) mensualidades, en caso de prórroga tácita o expresa, en forma automática y sin necesidad de requerimiento alguno entre las partes, el canon mensual del arrendamiento se incrementará en una proporción máxima a la fijada por el Gobierno Nacional para los casos de arriendo de vivienda urbana en el año que se produzca el reajuste, de acuerdo con lo establecido en el Artículo 20 de la ley 820 de julio de 2003.")

	addClause(pdf, "OCTAVA: PRORROGAS:",
		"Vencido el término pactado, si no existiere anuncio previo por las partes, el Contrato se entenderá prorrogado por seis (6) meses más. Si alguna de las partes no desea prorrogar el presente Contrato, tendrá que avisar a la otra con dos (2) meses de antelación al vencimiento del Contrato.")

	addClause(pdf, "NOVENA: SERVICIOS:",
		"Estarán a cargo del ARRENDATARIO el pago oportuno de los siguientes servicios públicos y privados: Energía Eléctrica, Acueducto y Alcantarillado, Gas, incluidos los servicios adicionales instalados bajo autorización y responsabilidad del ARRENDATARIO, previa autorización del ARRENDADOR. Los servicios de carácter privado, tales como antena parabólica, televisión satelital o por cable, Internet, Banda Ancha, avisos en páginas amarillas o cualquier otro, serán responsabilidad directa y exclusiva del ARRENDATARIO sobre lo cual EL ARRENDADOR no confiere autorización ni solidaridad alguna para su instalación y costos de servicio. PARAGRAFO PRIMERO. Las reclamaciones que tengan que ver con la óptima prestación o facturación de los servicios públicos anotados, serán tramitadas directamente por el ARRENDATARIO ante las respectivas empresas prestadoras del servicio. PARÁGRAFO SEGUNDO. Si el ARRENDATARIO no paga oportunamente los servicios públicos antes señalados, este hecho se tendrá como incumplimiento del contrato, pudiendo el ARRENDADOR darlo por terminado unilateralmente sin necesidad de los requerimientos privados y judiciales previstos en la Ley (artículos 1594 y 2007 del Código Civil). PARAGRAFO TERCERO. En cualquier evento de mora o retardo en el cumplimiento de las obligaciones a cargo del ARRENDATARIO, el ARRENDADOR queda facultado para exigir de aquel el pago de los honorarios de abogado y demás gastos de cobranza judicial y/o extrajudicial. Igualmente, sí como consecuencia del no pago oportuno de los servicios públicos las empresas respectivas los suspenden o retiran el contador, serán de cargo del ARRENDATARIO el pago de los intereses de mora, sanciones, y los gastos que demande su reconexión. PARÁGRAFO CUARTO. En caso de acordarse la prestación de garantías o fianzas por parte de el ARRENDATARIO, a favor de las entidades prestadoras de los servicios públicos antes indicadas, con el fin de garantizar a cada una de ellas el pago de las facturas correspondientes, tal pacto se hará constar en escrito separado, con el lleno de los requisitos exigidos para tal efecto. PARAGRAFO QUINTO. El presente documento junto con los recibos cancelados por el ARRENDADOR constituye título ejecutivo para cobrar judicialmente al ARRENDATARIO y sus garantes los servicios que dejaren de pagar siempre que tales montos correspondan al período en que éstos tuvieron en su poder los inmuebles. PARAGRAFO SEXTO. El ARRENDADOR no autoriza a el ARRENDATARIO para adquirir créditos, pólizas fúnebres, periódicos, revistas y/o periódicos, electrodomésticos y demás mediante las facturas de los servicios públicos.")

	addClause(pdf, "DECIMA: COSAS O USOS CONEXOS:",
		"Además de los inmuebles identificados y descritos anteriormente, tendrá el ARRENDATARIO derecho de goce de las zonas comunales de acuerdo con el manual de convivencia del Edificio y el Reglamento de Propiedad Horizontal. PARAGRAFO: El ARRENDATARIO declara que dentro de los inmuebles objeto del presente contrato está prohibido el uso, almacenamiento o consumo de sustancias prohibidas por la Ley.")

	addClause(pdf, "DECIMA PRIMERA: CLAUSULA PENAL:",
		"El incumplimiento por parte de EL ARRENDATARIO de cualquiera de las cláusulas de este contrato, y aún el simple retardo en el pago de una o más mensualidades y la evidente incursión en MORA y/o FALTA DE PAGO, lo constituirán en deudor del ARRENDADOR por una suma equivalente a dos (2) veces del canon mensual del arrendamiento que esté vigente en el momento en que tal incumplimiento se presente, a título de pena, que será exigible inmediatamente sin necesidad de requerimiento de ninguna clase y sin perjuicio de los demás derechos del ARRENDADOR, para hacer cesar el arrendamiento y exigir judicialmente la entrega de los inmuebles arrendados y el pago de la renta debida. Se entenderá en todo caso que el pago de la pena no extingue la obligación principal y que el ARRENDADOR podrá pedir a la vez el pago de la pena y la indemnización de los perjuicios, si es el caso. Este contrato será prueba sumaria suficiente para el cobro de esta pena y EL ARRENDATARIO renuncia expresamente a cualquier requerimiento privado o judicial para constituirlo en mora del pago de esta o de cualquier otra obligación derivada del contrato. PARÁGRAFO. Si el incumplimiento por parte del ARRENDATARIO fuere su deseo de dar por terminado el contrato en forma unilateral, antes del vencimiento inicial del mismo deberá pagar la indemnización prevista en el numeral 5 del artículo 24, de la Ley 820 de julio 10, 2003. De no mediar constancia por escrito del preaviso, el contrato de arrendamiento se entenderá renovado automáticamente por un término igual al inicialmente pactado.")

	addClause(pdf, "DECIMA SEGUNDA: REQUERIMIENTOS:",
		"El ARRENDATARIO renuncia expresamente a los requerimientos de tratan los artículos 2035 del C.C. y 424 del C. del P.C., relativos a la constitución en mora.")

	addClause(pdf, "DECIMA TERCERA: PREAVISO PARA LA ENTREGA:",
		"Las partes se obligan a dar el correspondiente preaviso para la entrega con dos (2) meses de anticipación al vencimiento del contrato. En todo caso, este preaviso deberá darse por escrito y a través de correo certificado o personalmente.")

	addClause(pdf, "DECIMA CUARTA: CESIÓN DE DERECHOS:",
		"Podrá el ARRENDADOR ceder libremente los derechos que emanan de este contrato y tal cesión producirá efectos respecto del ARRENDATARIO a partir de la fecha de la comunicación certificada en que a éstos se comunique.")

	addClause(pdf, "DECIMA QUINTA: CAUSALES DE TERMINACIÓN:",
		"A favor del ARRENDADOR serán las siguientes: a) La cesión del contrato o el subarriendo total o parcial de los inmuebles, b) El cambio de destinación de los inmuebles, c) El no pago del precio dentro del término previsto en este contrato, d) La destinación de los inmuebles para fines que afectan la tranquilidad ciudadana de los vecinos, o para fines ilícitos o contrarios a las buenas costumbres, o que representen peligro para los inmuebles o la salubridad de sus habitantes, e) La realización de mejoras, cambios o ampliaciones de los inmuebles, sin expresa autorización del ARRENDADOR. f) La no cancelación de los servicios públicos o privados (T.V., Internet, etc.) a cargo del ARRENDATARIO. g) La contravención o incumplimiento a cualquiera de las instrucciones o prohibiciones previstas con el presente contrato. i) Las demás previstas en la ley y en las cláusulas del presente contrato. A favor del ARRENDATARIO: a) La suspensión de la prestación de los servicios públicos al inmueble por acción o mora del ARRENDADOR. b) Los actos del ARRENDADOR que afecten gravemente el goce del bien arrendado. c) El desconocimiento por parte del ARRENDADOR de los derechos reconocidos al ARRENDATARIO por la Ley o el contrato. d) No recibir del ARRENDADOR copia del presente contrato, cuyas firmas sean originales, dentro de los diez (10) días siguientes a su celebración, so pena de hacerse acreedor a la sanción que a petición de parte imponga la autoridad competente, equivalente a una multa de una (1) mensualidad de arrendamiento. PARAGRAFO: Se tendrán como obligaciones especiales por parte del ARRENDADOR las siguientes: a) Mantener los inmuebles en buen estado de servir para el cumplimiento del objeto del contrato. b) Librar al ARRENDATARIO de toda turbación o desorden en el goce de los inmuebles. c) Hacer las reparaciones necesarias del bien objeto del arriendo, y las locativas, pero solo cuando estas provienen de fuerza mayor o caso fortuito, o de la mala calidad de la cosa arrendada.")

	addClause(pdf, "DECIMA SEXTA: RECIBO Y ESTADO:",
		"El ARRENDATARIO declara que ha recibido los inmuebles objeto de este contrato en buen estado, conforme al inventario que hace parte del mismo, y que en el mismo estado lo restituirá al ARRENDADOR a la terminación del contrato, o cuando éste haya de cesar por alguna de las causales previstas, salvo el deterioro proveniente del tiempo y uso legítimo. PARAGRAFO. No obstante, lo dispuesto en esta cláusula, el ARRENDATARIO está obligado a efectuar las reparaciones locativas para su conservación y cuidado conforme a los Artículos 2029 y 2030 del Código Civil. Las partes acuerdan un término, correspondiente a los quince (15) primeros días de vigencia del presente contrato para verificar el funcionamiento de todos y cada uno de los elementos y servicios de la vivienda, debiéndose dentro del mismo termino comunicar por escrito al ARRENDADOR las observaciones o deficiencias al respecto, para que igualmente sean verificadas y corregidas, según se hayan hecho constar como observaciones dentro del inventario referido. Los daños al inmueble derivados del mal trato o descuido por parte del ARRENDATARIO, durante su tenencia, serán de su cargo y el ARRENDADOR estará facultado para hacerlos por su cuenta y posteriormente reclamar su valor al ARRENDATARIO.")

	addClause(pdf, "DECIMA SEPTIMA.- RESTITUCION DE LOS INMUEBLES:",
		"Terminado el presente contrato, el ARRENDATARIO, deberá entregar los precitados Inmuebles al ARRENDADOR en forma personal o a quien éste autorice para recibirlo, conforme al Inventario inicial, obligándose a presentar los recibos de servicios públicos debidamente pagados. En relación con los servicios públicos pendientes de verificar, el ARRENDATARIO garantizará su pago mediante provisión proporcional y equivalente al promedio de sus tres (3) últimos consumos según la facturación respectiva. No será valida ni se entenderá como entrega formal y material de los inmuebles arrendados la que se realice por medios diferentes a los estipulados en la Ley o en el presente contrato. PARAGRAFO: El ARRENDATARIO se compromete girar o transferir al ARRENDADOR el valor de los servicios públicos que por su periodo de facturación correspondan al uso de los inmuebles por parte del ARRENDATARIO.")

	decimaOctavaText := fmt.Sprintf("El suscrito, %s con CC %s por medio del presente documento se declara deudor del ARRENDADOR en forma solidaria e indivisible junto con los ARRENDATARIOS de todas las cargas y obligaciones contenidas en el presente contrato, tanto durante el término inicialmente pactado como durante sus prórrogas o renovaciones expresas o tácitas y hasta la restitución real de los inmuebles al ARRENDADOR, por concepto de: Arrendamientos, servicios públicos. indemnizaciones, daños en los inmuebles, muebles y enseres,, cláusulas penales, costas procesales y cualquier otra derivada del contrato, las cuales podrán ser exigidas por el ARRENDADOR a cualquiera de los obligados, por la vía ejecutiva, sin necesidad de requerimientos privados o judiciales a los cuales renunciamos expresamente, sin que por razón de esta solidaridad asuma el carácter de fiador ni ARRENDATARIO del inmueble objeto del presente contrato, pues tal calidad la asume exclusivamente %s y sus respectivos causa-habitantes. Todo lo anterior sin perjuicio de que en caso de abandono de los inmuebles LOS DEUDORES SOLIDARIOS pueden hacer entrega válidamente de los mismos al ARRENDADOR o a quien éste señale, bien sea judicial o extrajudicialmente. Para este exclusivo efecto el ARRENDATARIO otorga poder amplio y suficiente a su DEUDOR SOLIDARIO en este mismo acto y al suscribir el presente contrato.", codeudorName, codeudorCC, arrendatarioName)

	addClause(pdf, "DECIMA OCTAVA: DEUDORES SOLIDARIOS.", decimaOctavaText)

	addClause(pdf, "DECIMA NOVENA: MEJORAS:",
		"No podrá el ARRENDATARIO ejecutar en los inmuebles mejoras de ninguna especie sin permiso escrito del ARRENDADOR, y si éstas se ejecutaren accederán al propietario de los inmuebles sin indemnización para quien las efectuó. No obstante, queda entendido que las reparaciones a cargo del ARRENDADOR serán todas aquellas contempladas en la Ley.")

	addClause(pdf, "VIGÉSIMA: MERITO EJECUTIVO DEL CONTRATO.",
		"Las partes acuerdan que el documento que contiene el presente contrato presta merito ejecutivo para efectos extrajudiciales y judiciales, con relación a todas las obligaciones que de éste se deriven, sin importar que la exigibilidad de las mismas, se haga con posterioridad a la restitución del inmueble. Los efectos del título ejecutivo se extenderán aún después de la restitución y hasta el cumplimiento total de las obligaciones a cargo del Arrendatario y del Deudor Solidario.")

	addClause(pdf, "VIGÉSIMA PRIMERA: VISITAS DE INSPECCION:",
		"El arrendador o su representante debidamente autorizado, está facultado para realizar una (1) visita al inmueble de ser necesario una vez dentro del plazo del contrato, con la finalidad de constatar la destinación, el estado y conservación del inmueble u otras circunstancias relacionadas con el contrato de arrendamiento, igualmente durante el preaviso legal para la terminación del contrato, acordando cita previa.")

	addClause(pdf, "VIGESIMA SEGUNDA: ABANDONO DE LOS INMUEBLES:",
		"Al suscribir este contrato el ARRENDATARIO faculta expresamente al ARRENDADOR para penetrar en los inmuebles y recuperar su tenencia, con el solo requisito de la presencia de dos (2) testigos, en procura de evitar el deterioro o el desmantelamiento de tal inmueble siempre que por cualquier circunstancia el mismo permanezca abandonado o deshabitado por el término de un mes o más y que la exposición al riesgo sea tal que amenace la integridad física del bien o la seguridad del vecindario.")

	addClause(pdf, "VIGESIMA TERCERA: NOTIFICACIONES JUDICIALES:",
		"En atención del articulo 103 del Código general del proceso, con el que se promueve el uso de las tecnologías de la información y de las comunicaciones bajo los principios de equivalencia funcional y neutralidad electrónica, así como del articulo 82 numeral 10 del mismo código en el que se tiene por requisito informar las direcciones electrónicas, las partes convienen que para efectos de notificaciones judiciales y extrajudiciales, relacionadas directa o indirectamente con el contrato de arrendamiento, las mismas serán remitidas a los siguientes correos electrónicos: El arrendador: <vickyderosas2003@hotmail.com> El arrendatario: <smotavitam@gmail.com> Testigo: <lau.co99@gmail.com> Deudor solidario: <nescool101@gmail.com>")

	// Final paragraph
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(0, 5, fixSpanishChars(fmt.Sprintf("Para constancia firmamos las partes y ante testigo el dia %s y declara el ARRENDATARIO que ha recibido la respectiva copia del presente contrato de acuerdo con la clausula decima quinta, previa autenticacion de firmas de las partes en Notaria. Para efecto de recibir notificaciones judiciales y extrajudiciales, las partes en cumplimiento del Art. 12 de la Ley 820 del 2003, a continuacion, y al suscribir este contrato proceden a indicar sus respectivos datos y direcciones.", currentDate)), "", "J", false)

	// Add signature tables
	addSignatureTables(pdf, arrendadorName, arrendadorCC, arrendatarioName, arrendatarioCC, testigoName, testigoCC, codeudorName, codeudorCC)

	// Return PDF as bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Helper functions
func getPropertyAddress(property *model.Property) string {
	if property != nil && property.Address != "" {
		return property.Address
	}
	return "Carrera 18b No 145 - 08 Apto 201"
}

func getGarageNumber(property *model.Property) string {
	if property != nil && property.AptNumber != "" {
		return property.AptNumber
	}
	return "22"
}

func getBuildingName(property *model.Property) string {
	// For now, using default. Could be extended to get from property details
	return "ACQUA 145"
}

func addInfoLine(pdf *gofpdf.Fpdf, label, value string) {
	pdf.SetFont("Arial", "B", 10)
	if label != "" {
		pdf.Cell(60, 6, fixSpanishChars(label))
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(130, 6, fixSpanishChars(value))
	} else {
		pdf.Cell(60, 6, "")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(130, 6, fixSpanishChars(value))
	}
	pdf.Ln(6)
}

func addClause(pdf *gofpdf.Fpdf, title string, content string) {
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 7, fixSpanishChars(title), "0", 0, "L", false, 0, "")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(0, 5, fixSpanishChars(content), "0", "J", false)
	pdf.Ln(3)
}

func addSignatureTables(pdf *gofpdf.Fpdf, arrendadorName, arrendadorCC, arrendatarioName, arrendatarioCC, testigoName, testigoCC, codeudorName, codeudorCC string) {
	pdf.Ln(10)

	// First table: ARRENDADOR and ARRENDATARIO
	pdf.SetFont("Arial", "B", 10)

	// Table headers
	cellWidth := 95.0
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("ARRENDADOR"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("ARRENDATARIO"), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// Names
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(cellWidth, 8, fixSpanishChars(arrendadorName), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars(arrendatarioName), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// CCs
	pdf.CellFormat(cellWidth, 8, "CC "+arrendadorCC, "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, "CC "+arrendatarioCC, "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// Addresses (using default values for now)
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("Dir. Notificacion: Calle 145 No."), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("Dir. Notificacion: Calle 167"), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	pdf.CellFormat(cellWidth, 8, fixSpanishChars("7F-60 Apto. 104"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("# 56 - 25 INT 2 APTO 103"), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// Emails
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("E-mail: vickyderosas2003@hotmail.com"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("E-mail: smotavitam@gmail.com"), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// Phones
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("Celular: 3204059245"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("Celular: 320 8692814"), "1", 0, "C", false, 0, "")
	pdf.Ln(15)

	// Second table: TESTIGO and CODEUDOR SOLIDARIO
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("TESTIGO"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("CODEUDOR SOLIDARIO"), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// Names
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(cellWidth, 8, fixSpanishChars(testigoName), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars(codeudorName), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// CCs
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("CC "+testigoCC), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("CC "+codeudorCC), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// Addresses
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("Dir. Notificacion: Calle 167 #"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("Dir. Notificacion: Calle 163"), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	pdf.CellFormat(cellWidth, 8, fixSpanishChars("56 - 25 INT 2 APTO 103"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("# 54 - 15, Casa 68"), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// Emails
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("E-mail: lau.co99@gmail.com"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("E-mail: nescool101@gmail.com"), "1", 0, "C", false, 0, "")
	pdf.Ln(8)

	// Phones
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("Celular: 3006631448"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(cellWidth, 8, fixSpanishChars("Celular: 3124894828"), "1", 0, "C", false, 0, "")
}

// FormatSpanishDateWithDay formats a date in Spanish format with day number in words
func FormatSpanishDateWithDay(date time.Time) string {
	if date.IsZero() {
		return "Fecha no especificada"
	}

	months := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio",
		"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}

	day := date.Day()
	dayStr := fmt.Sprintf("%s (%s)", NumberToWords(day), strings.ToUpper(NumberToWords(day)))

	return fmt.Sprintf("%s DE %s DEL AÑO %d", dayStr, strings.ToUpper(months[date.Month()-1]), date.Year())
}

// Convert a number to words in Spanish
func NumberToWords(n int) string {
	if n < 0 || n > 100 {
		return fmt.Sprintf("%d", n) // Return the number as is for numbers outside our range
	}

	// Basic numbers
	units := []string{"", "uno", "dos", "tres", "cuatro", "cinco", "seis", "siete", "ocho", "nueve",
		"diez", "once", "doce", "trece", "catorce", "quince", "dieciséis", "diecisiete", "dieciocho", "diecinueve"}
	tens := []string{"", "", "veinte", "treinta", "cuarenta", "cincuenta", "sesenta", "setenta", "ochenta", "noventa"}

	if n < 20 {
		return units[n]
	} else if n < 30 {
		if n == 20 {
			return tens[2]
		}
		return "veinti" + units[n-20]
	} else if n%10 == 0 {
		return tens[n/10]
	} else {
		return tens[n/10] + " y " + units[n%10]
	}
}

// AmountInWords converts a float64 amount to words in Spanish
func AmountInWords(amount float64) string {
	// Round to whole pesos
	intAmount := int(amount)

	if intAmount == 0 {
		return "CERO"
	}

	// For simplicity, we'll only handle amounts up to 9,999,999
	if intAmount > 9999999 {
		return fmt.Sprintf("%d", intAmount)
	}

	millions := intAmount / 1000000
	thousands := (intAmount % 1000000) / 1000
	units := intAmount % 1000

	words := ""

	if millions > 0 {
		if millions == 1 {
			words += "UN MILLÓN "
		} else {
			words += strings.ToUpper(NumberToWords(millions)) + " MILLONES "
		}
	}

	if thousands > 0 {
		if thousands == 1 {
			words += "MIL "
		} else {
			words += strings.ToUpper(NumberToWords(thousands)) + " MIL "
		}
	}

	if units > 0 {
		words += strings.ToUpper(NumberToWords(units))
	}

	return strings.TrimSpace(words)
}

// FormatMoney formats a number as a money string with commas and decimals
func FormatMoney(amount float64) string {
	if amount == 0 {
		return "$0.00"
	}

	// Format with two decimal places
	formattedAmount := fmt.Sprintf("$%.2f", amount)

	// Split the string at the decimal point
	parts := strings.Split(formattedAmount, ".")

	// Get the integer part
	intPart := parts[0][1:] // Remove the dollar sign

	// Format with commas
	var result string
	for i, j := len(intPart)-1, 0; i >= 0; i, j = i-1, j+1 {
		if j > 0 && j%3 == 0 {
			result = "," + result
		}
		result = string(intPart[i]) + result
	}

	// Add the dollar sign and decimal part back
	return "$" + result + "." + parts[1]
}

// FormatSpanishDate formats a date in Spanish format
func FormatSpanishDate(date time.Time) string {
	if date.IsZero() {
		return "Fecha no especificada"
	}

	months := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio",
		"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}

	return fmt.Sprintf("%d de %s de %d", date.Day(), months[date.Month()-1], date.Year())
}

// fixSpanishChars converts problematic Spanish characters to properly display in PDF
func fixSpanishChars(text string) string {
	// Convert common Spanish accented characters that might cause encoding issues
	replacements := map[string]string{
		"á": "a", "é": "e", "í": "i", "ó": "o", "ú": "u",
		"Á": "A", "É": "E", "Í": "I", "Ó": "O", "Ú": "U",
		"ñ": "n", "Ñ": "N",
		"ü": "u", "Ü": "U",
	}

	result := text
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}
