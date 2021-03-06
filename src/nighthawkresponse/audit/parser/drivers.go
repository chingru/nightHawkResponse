package parser

import (
	"encoding/xml"
	"fmt"
	"os"

	nhc "nighthawkresponse/common"
	nhconfig "nighthawkresponse/config"
	"nighthawkresponse/elastic"
	nhlog "nighthawkresponse/log"
	nhs "nighthawkresponse/nhstruct"
)

func ParseDriverModules(caseinfo nhs.CaseInformation, auditinfo nhs.AuditType, auditfile string) {
	MAX_RECORD := nhconfig.BulkPostSize()

	xmlFile, err := os.Open(auditfile)
	if err != nil {
		nhlog.LogMessage("ParseDriverModules", "ERROR", fmt.Sprintf("Failed to read audit file %s", err.Error()))
		os.Exit(nhc.ERROR_READING_AUDIT_FILE)
	}
	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)
	count := 0
	total := 0

	var inElement string
	var esrecords []nhs.RlRecord

	for {
		if count == MAX_RECORD {
			elastic.ProcessOutput(caseinfo, auditinfo, esrecords)
			esrecords = esrecords[:0]
			count = 0
		}

		t, _ := decoder.Token()
		if t == nil {
			elastic.ProcessOutput(caseinfo, auditinfo, esrecords)
			esrecords = esrecords[:0]
			count++

			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			inElement = se.Name.Local

			if inElement == "ModuleItem" {
				var item nhs.ModuleItem
				decoder.DecodeElement(&item, &se)
				item.IsWhitelisted = false

				var rlrec nhs.RlRecord
				rlrec.ComputerName = caseinfo.ComputerName
				rlrec.CaseInfo = caseinfo
				rlrec.AuditType = auditinfo
				rlrec.Record = item

				esrecords = append(esrecords, rlrec)
				count++
				total++
			}
		} // __end_switch__
	} // __end_for__

	cmsg := fmt.Sprintf("Total DriverModules %d", total)
	nhlog.LogMessage("ParseDriverModules", "INFO", cmsg)
}
