// Package ods exports TTT ledger state as OpenDocument spreadsheets.
package ods

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"ttt/engine"
)

const odsMimetype = "application/vnd.oasis.opendocument.spreadsheet"

// ExportStore writes an ODS workbook for all accounts and journal lines in store.
func ExportStore(store engine.Store, path string) error {
	accounts, err := store.ListAccounts()
	if err != nil {
		return fmt.Errorf("list accounts: %w", err)
	}
	lines, err := store.GetAllLines()
	if err != nil {
		return fmt.Errorf("list journal lines: %w", err)
	}
	return Export(path, accounts, lines)
}

// Export writes an ODS workbook for the supplied accounts and journal lines.
func Export(path string, accounts []engine.Account, lines []engine.JournalLine) error {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	if err := addMimetype(zw); err != nil {
		return err
	}
	files := map[string]string{
		"META-INF/manifest.xml": manifestXML(),
		"content.xml":           contentXML(accounts, lines),
		"styles.xml":            stylesXML(),
		"meta.xml":              metaXML(),
	}
	for name, body := range files {
		w, err := createZipFile(zw, name, zip.Deflate)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, body); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func addMimetype(zw *zip.Writer) error {
	h := &zip.FileHeader{
		Name:               "mimetype",
		Method:             zip.Store,
		CRC32:              crc32.ChecksumIEEE([]byte(odsMimetype)),
		CompressedSize:     uint32(len(odsMimetype)),
		CompressedSize64:   uint64(len(odsMimetype)),
		UncompressedSize:   uint32(len(odsMimetype)),
		UncompressedSize64: uint64(len(odsMimetype)),
		ModifiedDate:       dosDate(1980, 1, 1),
		ModifiedTime:       dosTime(0, 0, 0),
	}
	w, err := zw.CreateRaw(h)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, odsMimetype)
	return err
}

func createZipFile(zw *zip.Writer, name string, method uint16) (io.Writer, error) {
	h := &zip.FileHeader{
		Name:         name,
		Method:       method,
		ModifiedDate: dosDate(1980, 1, 1),
		ModifiedTime: dosTime(0, 0, 0),
	}
	h.SetMode(0o644)
	return zw.CreateHeader(h)
}

func dosDate(year, month, day int) uint16 {
	return uint16((year-1980)<<9 | month<<5 | day)
}

func dosTime(hour, minute, second int) uint16 {
	return uint16(hour<<11 | minute<<5 | second/2)
}

func contentXML(accounts []engine.Account, lines []engine.JournalLine) string {
	activeAccounts := activeAccountColumns(accounts, lines)
	accountGroups := providerGroups(activeAccounts)
	accountTotals := totalsByAccount(activeAccounts, lines)
	accountByID := make(map[string]engine.Account, len(accounts))
	for _, a := range accounts {
		accountByID[a.ID] = a
	}

	var b strings.Builder
	b.WriteString(xml.Header)
	b.WriteString(`<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0" xmlns:of="urn:oasis:names:tc:opendocument:xmlns:of:1.2" xmlns:number="urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0" office:version="1.2" office:mimetype="application/vnd.oasis.opendocument.spreadsheet">`)
	b.WriteString(automaticStylesXML())
	b.WriteString(`<office:body><office:spreadsheet><table:table table:name="Scenario">`)
	b.WriteString(`<table:table-column table:number-columns-repeated="1" table:style-name="col-small"/>`)
	b.WriteString(`<table:table-column table:number-columns-repeated="2" table:style-name="col-medium"/>`)
	b.WriteString(`<table:table-column table:number-columns-repeated="5" table:style-name="col-wide"/>`)
	if len(activeAccounts) > 0 {
		b.WriteString(fmt.Sprintf(`<table:table-column table:number-columns-repeated="%d" table:style-name="col-money"/>`, len(activeAccounts)*2))
	}

	writeGroupHeaderRow(&b, accountGroups)

	headers := []odsCell{
		styledTextCell("#", "hdr-meta"),
		styledTextCell("Timestamp", "hdr-meta"),
		styledTextCell("Event", "hdr-meta"),
		styledTextCell("Workflow", "hdr-meta"),
		styledTextCell("Step", "hdr-meta"),
		styledTextCell("Debit Account", "hdr-debit"),
		styledTextCell("Credit Account", "hdr-credit"),
		styledTextCell("FX Rate", "hdr-fx"),
	}
	for _, a := range activeAccounts {
		headers = append(headers, mergedTextCell(accountShortLabelForODS(a), "hdr-account", 2), coveredCell())
	}
	writeRow(&b, headers...)

	sideHeaders := []odsCell{
		emptyStyledCell("hdr-meta"),
		emptyStyledCell("hdr-meta"),
		emptyStyledCell("hdr-meta"),
		emptyStyledCell("hdr-meta"),
		emptyStyledCell("hdr-meta"),
		emptyStyledCell("hdr-debit"),
		emptyStyledCell("hdr-credit"),
		emptyStyledCell("hdr-fx"),
	}
	for range activeAccounts {
		sideHeaders = append(sideHeaders, styledTextCell("Debit", "hdr-debit"), styledTextCell("Credit", "hdr-credit"))
	}
	writeRow(&b, sideHeaders...)

	for i, line := range lines {
		debitAccount := accountByID[line.DebitAccountID]
		creditAccount := accountByID[line.CreditAccountID]
		cells := []odsCell{
			styledNumberCell(int64(i+1), "cell-meta"),
			styledTextCell(line.Timestamp.Format(time.RFC3339), "cell-meta"),
			styledTextCell(shortID(line.EventID), "cell-meta"),
			styledTextCell(line.Metadata[engine.MetaWorkflow], "cell-meta"),
			styledTextCell(stepText(line), "cell-meta"),
			styledTextCell(accountLabelForODS(debitAccount), "cell-debit"),
			styledTextCell(accountLabelForODS(creditAccount), "cell-credit"),
			fxRateCell(line),
		}
		for _, a := range activeAccounts {
			if line.DebitAccountID == a.ID {
				cells = append(cells, styledAmountCell(line.Amount, a.Currency.AssetScale, currencyCellStyle("amt-debit", a.Currency.Code)))
			} else {
				cells = append(cells, emptyStyledCell("amt-empty"))
			}
			if line.CreditAccountID == a.ID {
				cells = append(cells, styledAmountCell(line.Amount, a.Currency.AssetScale, currencyCellStyle("amt-credit", a.Currency.Code)))
			} else {
				cells = append(cells, emptyStyledCell("amt-empty"))
			}
		}
		writeRow(&b, cells...)
	}

	firstDataRow := 4
	lastDataRow := len(lines) + 3
	totalsRow := len(lines) + 4
	globalRow := len(lines) + 7

	totalCells := []odsCell{styledTextCell("Totals", "summary-label"), emptyCell(), emptyCell(), emptyCell(), emptyCell(), emptyCell(), emptyCell(), emptyCell()}
	for i := range activeAccounts {
		debitCol := accountDebitCol(i)
		creditCol := accountCreditCol(i)
		currencyCode := activeAccounts[i].Currency.Code
		totalCells = append(totalCells,
			styledFormulaCell(sumFormula(debitCol, firstDataRow, lastDataRow), accountTotals[i].debitDisplay(), currencyCellStyle("summary-debit", currencyCode)),
			styledFormulaCell(sumFormula(creditCol, firstDataRow, lastDataRow), accountTotals[i].creditDisplay(), currencyCellStyle("summary-credit", currencyCode)),
		)
	}
	writeRow(&b, totalCells...)

	balanceCells := []odsCell{styledTextCell("Account balance (credit - debit)", "summary-label"), emptyCell(), emptyCell(), emptyCell(), emptyCell(), emptyCell(), emptyCell(), emptyCell()}
	for i := range activeAccounts {
		debitRef := cellRef(accountDebitCol(i), totalsRow)
		creditRef := cellRef(accountCreditCol(i), totalsRow)
		balanceCells = append(
			balanceCells,
			emptyStyledCell("summary-empty"),
			styledFormulaCell(
				fmt.Sprintf("of:=%s-%s", creditRef, debitRef),
				accountTotals[i].balanceDisplay(),
				currencyCellStyle("summary-balance", activeAccounts[i].Currency.Code),
			),
		)
	}
	writeRow(&b, balanceCells...)
	writeRow(&b)

	globalDebitParts := make([]string, 0, len(activeAccounts))
	globalCreditParts := make([]string, 0, len(activeAccounts))
	globalDebit := 0.0
	globalCredit := 0.0
	for i := range activeAccounts {
		globalDebitParts = append(globalDebitParts, cellRef(accountDebitCol(i), totalsRow))
		globalCreditParts = append(globalCreditParts, cellRef(accountCreditCol(i), totalsRow))
		globalDebit += accountTotals[i].debitFloat()
		globalCredit += accountTotals[i].creditFloat()
	}
	writeRow(&b,
		styledTextCell("Global debit total", "summary-label"),
		styledFormulaCell(sumRefsFormula(globalDebitParts), fmtFloat(globalDebit), "summary-debit"),
		styledTextCell("Global credit total", "summary-label"),
		styledFormulaCell(sumRefsFormula(globalCreditParts), fmtFloat(globalCredit), "summary-credit"),
		styledTextCell("Difference", "summary-label"),
		styledFormulaCell(fmt.Sprintf("of:=%s-%s", cellRef("B", globalRow), cellRef("D", globalRow)), fmtFloat(globalDebit-globalCredit), "summary-balance"),
		styledTextCell("Balanced when difference is 0", "summary-label"),
	)

	b.WriteString(`</table:table></office:spreadsheet></office:body></office:document-content>`)
	return b.String()
}

type accountTotal struct {
	scale  int
	debit  int64
	credit int64
}

func totalsByAccount(accounts []engine.Account, lines []engine.JournalLine) []accountTotal {
	index := make(map[string]int, len(accounts))
	totals := make([]accountTotal, len(accounts))
	for i, a := range accounts {
		index[a.ID] = i
		totals[i].scale = a.Currency.AssetScale
	}
	for _, line := range lines {
		if i, ok := index[line.DebitAccountID]; ok {
			totals[i].debit += line.Amount
		}
		if i, ok := index[line.CreditAccountID]; ok {
			totals[i].credit += line.Amount
		}
	}
	return totals
}

func (t accountTotal) debitDisplay() string {
	return fmtScaledDecimal(t.debit, t.scale)
}

func (t accountTotal) creditDisplay() string {
	return fmtScaledDecimal(t.credit, t.scale)
}

func (t accountTotal) balanceDisplay() string {
	return fmtScaledDecimal(t.credit-t.debit, t.scale)
}

func (t accountTotal) debitFloat() float64 {
	return scaledFloat(t.debit, t.scale)
}

func (t accountTotal) creditFloat() float64 {
	return scaledFloat(t.credit, t.scale)
}

func scaledFloat(amount int64, scale int) float64 {
	divisor := float64(1)
	for i := 0; i < scale; i++ {
		divisor *= 10
	}
	return float64(amount) / divisor
}

func fmtFloat(v float64) string {
	if v == 0 {
		return "0"
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.8f", v), "0"), ".")
}

func activeAccountColumns(accounts []engine.Account, lines []engine.JournalLine) []engine.Account {
	active := make(map[string]bool)
	for _, line := range lines {
		active[line.DebitAccountID] = true
		active[line.CreditAccountID] = true
	}
	var out []engine.Account
	for _, a := range accounts {
		if active[a.ID] {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ProviderID != out[j].ProviderID {
			return out[i].ProviderID < out[j].ProviderID
		}
		if out[i].Currency.Code != out[j].Currency.Code {
			return out[i].Currency.Code < out[j].Currency.Code
		}
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		return accountLabelForODS(out[i]) < accountLabelForODS(out[j])
	})
	return out
}

type accountGroup struct {
	providerID string
	count      int
}

func providerGroups(accounts []engine.Account) []accountGroup {
	if len(accounts) == 0 {
		return nil
	}
	groups := []accountGroup{{providerID: accounts[0].ProviderID, count: 1}}
	for _, a := range accounts[1:] {
		last := &groups[len(groups)-1]
		if a.ProviderID == last.providerID {
			last.count++
			continue
		}
		groups = append(groups, accountGroup{providerID: a.ProviderID, count: 1})
	}
	return groups
}

func writeGroupHeaderRow(b *strings.Builder, groups []accountGroup) {
	cells := []odsCell{mergedTextCell("Event details", "group-meta", leadingColumns)}
	for i := 1; i < leadingColumns; i++ {
		cells = append(cells, coveredCell())
	}
	for i, group := range groups {
		style := "group-provider-a"
		if i%2 == 1 {
			style = "group-provider-b"
		}
		span := group.count * 2
		cells = append(cells, mergedTextCell(group.providerID, style, span))
		for j := 1; j < span; j++ {
			cells = append(cells, coveredCell())
		}
	}
	writeRow(b, cells...)
}

type odsCell struct {
	valueType string
	value     string
	formula   string
	text      string
	style     string
	colSpan   int
	covered   bool
}

func textCell(s string) odsCell {
	return odsCell{valueType: "string", text: s}
}

func styledTextCell(s, style string) odsCell {
	return odsCell{valueType: "string", text: s, style: style}
}

func mergedTextCell(s, style string, colSpan int) odsCell {
	if colSpan < 1 {
		colSpan = 1
	}
	return odsCell{valueType: "string", text: s, style: style, colSpan: colSpan}
}

func numberCell(n int64) odsCell {
	return odsCell{valueType: "float", value: fmt.Sprintf("%d", n), text: fmt.Sprintf("%d", n)}
}

func styledNumberCell(n int64, style string) odsCell {
	return odsCell{valueType: "float", value: fmt.Sprintf("%d", n), text: fmt.Sprintf("%d", n), style: style}
}

func amountCell(amount int64, scale int) odsCell {
	value := fmtScaledDecimal(amount, scale)
	return odsCell{valueType: "float", value: value, text: value}
}

func styledAmountCell(amount int64, scale int, style string) odsCell {
	value := fmtScaledDecimal(amount, scale)
	return odsCell{valueType: "float", value: value, text: value, style: style}
}

func formulaCell(formula string) odsCell {
	return odsCell{valueType: "float", value: "0", formula: formula, text: "0"}
}

func styledFormulaCell(formula, value, style string) odsCell {
	return odsCell{valueType: "float", value: value, formula: formula, text: value, style: style}
}

func currencyCellStyle(base, code string) string {
	return base + "-" + strings.ToLower(code)
}

func emptyCell() odsCell {
	return odsCell{}
}

func emptyStyledCell(style string) odsCell {
	return odsCell{style: style}
}

func coveredCell() odsCell {
	return odsCell{covered: true}
}

func stringCells(values []string) []odsCell {
	cells := make([]odsCell, 0, len(values))
	for _, v := range values {
		cells = append(cells, textCell(v))
	}
	return cells
}

func writeRow(b *strings.Builder, cells ...odsCell) {
	b.WriteString(`<table:table-row>`)
	for _, c := range cells {
		writeCell(b, c)
	}
	b.WriteString(`</table:table-row>`)
}

func writeCell(b *strings.Builder, c odsCell) {
	if c.covered {
		b.WriteString(`<table:covered-table-cell/>`)
		return
	}
	if c.valueType == "" {
		b.WriteString(`<table:table-cell`)
		writeCellAttrs(b, c)
		b.WriteString(`/>`)
		return
	}
	b.WriteString(`<table:table-cell office:value-type="`)
	b.WriteString(c.valueType)
	b.WriteString(`"`)
	writeCellAttrs(b, c)
	if c.value != "" {
		b.WriteString(` office:value="`)
		b.WriteString(esc(c.value))
		b.WriteString(`"`)
	}
	if c.formula != "" {
		b.WriteString(` table:formula="`)
		b.WriteString(esc(c.formula))
		b.WriteString(`"`)
	}
	b.WriteString(`><text:p>`)
	b.WriteString(esc(c.text))
	b.WriteString(`</text:p></table:table-cell>`)
}

func writeCellAttrs(b *strings.Builder, c odsCell) {
	if c.style != "" {
		b.WriteString(` table:style-name="`)
		b.WriteString(esc(c.style))
		b.WriteString(`"`)
	}
	if c.colSpan > 1 {
		b.WriteString(` table:number-columns-spanned="`)
		b.WriteString(fmt.Sprintf("%d", c.colSpan))
		b.WriteString(`"`)
	}
}

const leadingColumns = 8

func accountDebitCol(accountIndex int) string {
	return spreadsheetCol(leadingColumns + 1 + accountIndex*2)
}

func accountCreditCol(accountIndex int) string {
	return spreadsheetCol(leadingColumns + 2 + accountIndex*2)
}

func sumFormula(col string, firstRow, lastRow int) string {
	if lastRow < firstRow {
		return "of:=0"
	}
	return fmt.Sprintf("of:=SUM([.%s%d:.%s%d])", col, firstRow, col, lastRow)
}

func sumRefsFormula(refs []string) string {
	if len(refs) == 0 {
		return "of:=0"
	}
	return "of:=SUM(" + strings.Join(refs, ";") + ")"
}

func cellRef(col string, row int) string {
	return fmt.Sprintf("[.%s%d]", col, row)
}

func spreadsheetCol(n int) string {
	var out []byte
	for n > 0 {
		n--
		out = append([]byte{byte('A' + n%26)}, out...)
		n /= 26
	}
	return string(out)
}

func stepText(line engine.JournalLine) string {
	if step := line.Metadata[engine.MetaStep]; step != "" {
		return step
	}
	if step := line.DebitMetadata[engine.MetaStep]; step != "" {
		if credit := line.CreditMetadata[engine.MetaStep]; credit != "" && credit != step {
			return step + " / " + credit
		}
		return step
	}
	return line.CreditMetadata[engine.MetaStep]
}

func fxRateCell(line engine.JournalLine) odsCell {
	numText := line.Metadata[engine.MetaFXRateNum]
	denText := line.Metadata[engine.MetaFXRateDen]
	if numText == "" || denText == "" {
		return emptyStyledCell("cell-fx")
	}
	num, err := strconv.ParseFloat(numText, 64)
	if err != nil {
		return styledTextCell(numText+"/"+denText, "cell-fx")
	}
	den, err := strconv.ParseFloat(denText, 64)
	if err != nil || den == 0 {
		return styledTextCell(numText+"/"+denText, "cell-fx")
	}
	pair := strings.TrimSpace(line.Metadata[engine.MetaFXBase] + "/" + line.Metadata[engine.MetaFXQuote])
	if pair == "/" {
		pair = ""
	}
	text := fmt.Sprintf("%.6f", num/den)
	if pair != "" {
		text += " " + pair
	}
	return styledTextCell(text, "cell-fx")
}

func accountLabelForODS(a engine.Account) string {
	switch a.Type {
	case engine.AccountTypeSystem:
		return fmt.Sprintf("%s/system(%s)", a.ProviderID, a.Currency.Code)
	case engine.AccountTypeLiquidity:
		return fmt.Sprintf("%s/liquidity(%s)", a.ProviderID, a.Currency.Code)
	case engine.AccountTypePosition:
		return fmt.Sprintf("%s/position(%s)@%s", a.ProviderID, a.Currency.Code, a.CounterpartyID)
	case engine.AccountTypeUser:
		return fmt.Sprintf("%s/%s(%s)", a.ProviderID, a.UserID, a.Currency.Code)
	case engine.AccountTypeFX:
		return fmt.Sprintf("%s/fx(%s)", a.ProviderID, a.Currency.Code)
	default:
		if a.ID == "" {
			return "(unknown)"
		}
		return a.ID
	}
}

func accountShortLabelForODS(a engine.Account) string {
	switch a.Type {
	case engine.AccountTypeSystem:
		return fmt.Sprintf("system %s", a.Currency.Code)
	case engine.AccountTypeLiquidity:
		return fmt.Sprintf("liquidity %s", a.Currency.Code)
	case engine.AccountTypePosition:
		return fmt.Sprintf("position %s @ %s", a.Currency.Code, a.CounterpartyID)
	case engine.AccountTypeUser:
		return fmt.Sprintf("%s %s", a.UserID, a.Currency.Code)
	case engine.AccountTypeFX:
		return fmt.Sprintf("fx %s", a.Currency.Code)
	default:
		return accountLabelForODS(a)
	}
}

func fmtScaledDecimal(amount int64, scale int) string {
	if scale <= 0 {
		return fmt.Sprintf("%d", amount)
	}
	neg := amount < 0
	if neg {
		amount = -amount
	}
	pow := int64(1)
	for i := 0; i < scale; i++ {
		pow *= 10
	}
	out := fmt.Sprintf("%d.%0*d", amount/pow, scale, amount%pow)
	if neg {
		return "-" + out
	}
	return out
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

var slugRE = regexp.MustCompile(`[^a-z0-9]+`)

// Slug returns a filesystem-friendly lowercase slug.
func Slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRE.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "scenario"
	}
	return s
}

func esc(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func manifestXML() string {
	return xml.Header + `<manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2"><manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.spreadsheet"/><manifest:file-entry manifest:full-path="mimetype" manifest:media-type="text/plain"/><manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/><manifest:file-entry manifest:full-path="styles.xml" manifest:media-type="text/xml"/><manifest:file-entry manifest:full-path="meta.xml" manifest:media-type="text/xml"/></manifest:manifest>`
}

func automaticStylesXML() string {
	return `<office:scripts/><office:font-face-decls/><office:automatic-styles>` +
		currencyDataStyleXML("cur-eur", "EUR") +
		currencyDataStyleXML("cur-zar", "ZAR") +
		columnStyleXML("col-small", "0.55in") +
		columnStyleXML("col-medium", "1.25in") +
		columnStyleXML("col-wide", "2.25in") +
		columnStyleXML("col-money", "1.05in") +
		cellStyleXML("group-meta", "#263238", "#FFFFFF", true, "center") +
		cellStyleXML("group-provider-a", "#0D47A1", "#FFFFFF", true, "center") +
		cellStyleXML("group-provider-b", "#2E7D32", "#FFFFFF", true, "center") +
		cellStyleXML("hdr-meta", "#ECEFF1", "#263238", true, "center") +
		cellStyleXML("hdr-account", "#E3F2FD", "#0D47A1", true, "center") +
		cellStyleXML("hdr-debit", "#FFEBEE", "#B71C1C", true, "center") +
		cellStyleXML("hdr-credit", "#E8F5E9", "#1B5E20", true, "center") +
		cellStyleXML("hdr-fx", "#EDE7F6", "#4527A0", true, "center") +
		cellStyleXML("cell-meta", "#FFFFFF", "#263238", false, "start") +
		cellStyleXML("cell-debit", "#FFF7F7", "#263238", false, "start") +
		cellStyleXML("cell-credit", "#F7FFF8", "#263238", false, "start") +
		cellStyleXML("cell-fx", "#F5F0FF", "#4527A0", false, "end") +
		cellStyleXML("amt-debit", "#FFEBEE", "#B71C1C", false, "end") +
		cellStyleDataXML("amt-debit-eur", "#FFEBEE", "#B71C1C", false, "end", "cur-eur") +
		cellStyleDataXML("amt-debit-zar", "#FFEBEE", "#B71C1C", false, "end", "cur-zar") +
		cellStyleXML("amt-credit", "#E8F5E9", "#1B5E20", false, "end") +
		cellStyleDataXML("amt-credit-eur", "#E8F5E9", "#1B5E20", false, "end", "cur-eur") +
		cellStyleDataXML("amt-credit-zar", "#E8F5E9", "#1B5E20", false, "end", "cur-zar") +
		cellStyleXML("amt-empty", "#FAFAFA", "#263238", false, "end") +
		cellStyleXML("summary-label", "#FFF8E1", "#3E2723", true, "start") +
		cellStyleXML("summary-debit", "#FFCDD2", "#B71C1C", true, "end") +
		cellStyleDataXML("summary-debit-eur", "#FFCDD2", "#B71C1C", true, "end", "cur-eur") +
		cellStyleDataXML("summary-debit-zar", "#FFCDD2", "#B71C1C", true, "end", "cur-zar") +
		cellStyleXML("summary-credit", "#C8E6C9", "#1B5E20", true, "end") +
		cellStyleDataXML("summary-credit-eur", "#C8E6C9", "#1B5E20", true, "end", "cur-eur") +
		cellStyleDataXML("summary-credit-zar", "#C8E6C9", "#1B5E20", true, "end", "cur-zar") +
		cellStyleXML("summary-balance", "#BBDEFB", "#0D47A1", true, "end") +
		cellStyleDataXML("summary-balance-eur", "#BBDEFB", "#0D47A1", true, "end", "cur-eur") +
		cellStyleDataXML("summary-balance-zar", "#BBDEFB", "#0D47A1", true, "end", "cur-zar") +
		cellStyleXML("summary-empty", "#FFF8E1", "#3E2723", false, "end") +
		`</office:automatic-styles>`
}

func currencyDataStyleXML(name, code string) string {
	return fmt.Sprintf(`<number:currency-style style:name="%s"><number:number number:decimal-places="2" number:min-decimal-places="2" number:grouping="true"/><number:text> </number:text><number:currency-symbol>%s</number:currency-symbol></number:currency-style>`, name, code)
}

func columnStyleXML(name, width string) string {
	return fmt.Sprintf(`<style:style style:name="%s" style:family="table-column"><style:table-column-properties style:column-width="%s"/></style:style>`, name, width)
}

func cellStyleXML(name, bg, fg string, bold bool, align string) string {
	return cellStyleDataXML(name, bg, fg, bold, align, "")
}

func cellStyleDataXML(name, bg, fg string, bold bool, align, dataStyle string) string {
	weight := "normal"
	if bold {
		weight = "bold"
	}
	dataAttr := ""
	if dataStyle != "" {
		dataAttr = fmt.Sprintf(` style:data-style-name="%s"`, dataStyle)
	}
	return fmt.Sprintf(`<style:style style:name="%s" style:family="table-cell"%s><style:table-cell-properties fo:background-color="%s" fo:border="0.5pt solid #CFD8DC" fo:padding="0.03in" fo:wrap-option="wrap"/><style:text-properties fo:color="%s" fo:font-weight="%s"/><style:paragraph-properties fo:text-align="%s"/></style:style>`, name, dataAttr, bg, fg, weight, align)
}

func stylesXML() string {
	return xml.Header + `<office:document-styles xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" office:version="1.2"><office:styles/></office:document-styles>`
}

func metaXML() string {
	return xml.Header + `<office:document-meta xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0" office:version="1.2"><office:meta><meta:generator>TTT e2e</meta:generator></office:meta></office:document-meta>`
}
