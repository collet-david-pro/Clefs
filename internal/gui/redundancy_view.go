package gui

import (
	"clefs/internal/db"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// createRedundancyView affiche les détenteurs ayant des accès couverts par plusieurs clés.
func createRedundancyView(a *App) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Redondances d'accès", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabel("Détenteurs dont plusieurs clés ouvrent les mêmes portes.")

	reports, err := db.GetBorrowersWithRedundantAccesses()
	if err != nil {
		return container.NewVBox(title, widget.NewLabel(fmt.Sprintf("Erreur : %v", err)))
	}

	list := container.NewVBox()

	if len(reports) == 0 {
		list.Add(widget.NewCard("", "Aucune redondance détectée",
			widget.NewLabel("Tous les détenteurs ont des accès distincts pour chacune de leurs clés.")))
	} else {
		for _, r := range reports {
			r := r

			keyNames := make([]string, len(r.Keys))
			for i, k := range r.Keys {
				keyNames[i] = k.Number
			}
			redNames := make([]string, len(r.RedundantAccesses))
			for i, acc := range r.RedundantAccesses {
				redNames[i] = acc.Name
			}

			headerText := fmt.Sprintf("⚠️ %s — clés : %s", r.Borrower.Name, strings.Join(keyNames, ", "))
			detail := container.NewVBox(
				widget.NewLabel("Accès en doublon :"),
				widget.NewLabelWithStyle(
					strings.Join(redNames, ", "),
					fyne.TextAlignLeading,
					fyne.TextStyle{Bold: true},
				),
				widget.NewLabel("Ces accès sont couverts par plusieurs des clés détenues."),
			)

			list.Add(widget.NewAccordion(widget.NewAccordionItem(headerText, detail)))
		}
	}

	return container.NewBorder(
		container.NewVBox(title, subtitle, widget.NewSeparator()),
		nil, nil, nil,
		container.NewVScroll(list),
	)
}
