package pages

import (
	"encoding/json"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	gosxnotifier "github.com/deckarep/gosx-notifier"
	uniswap "github.com/hirokimoto/uniswap-api"
	unitrade "github.com/hirokimoto/uniswap-api/swap"
	unitrades "github.com/hirokimoto/uniswap-api/swaps"
	"github.com/uniswap-auto-gui/data"
	"github.com/uniswap-auto-gui/services"
)

func trackScreen(_ fyne.Window) fyne.CanvasObject {
	var selected uniswap.Swaps
	var activePair string
	var oldPrices = map[string]float64{}

	// money := accounting.Accounting{Symbol: "$", Precision: 6}
	trades := map[string]uniswap.Swaps{}

	pairs := data.ReadTrackPairs()

	name := widget.NewEntry()
	name.SetPlaceHolder("0x385769E84B650C070964398929DB67250B7ff72C")
	append := widget.NewButton("Append", func() {
		if name.Text != "" {
			isExisted := false
			for _, item := range pairs {
				if item == name.Text {
					isExisted = true
				}
			}
			if !isExisted {
				pairs = append(pairs, name.Text)
				data.SaveTrackPairs(pairs)
			}
		}
	})

	control := container.NewVBox(name, append)

	rightList := widget.NewList(
		func() int {
			return len(selected.Data.Swaps)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("target"), widget.NewLabel("price"), widget.NewLabel("amount"), widget.NewLabel("amount1"), widget.NewLabel("amount2"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			price, target, amount, amount1, amount2 := unitrade.Trade(selected.Data.Swaps[id])

			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(target)
			item.(*fyne.Container).Objects[2].(*widget.Label).SetText(fmt.Sprintf("$%f", price))
			item.(*fyne.Container).Objects[3].(*widget.Label).SetText(amount)
			item.(*fyne.Container).Objects[4].(*widget.Label).SetText(amount1)
			item.(*fyne.Container).Objects[5].(*widget.Label).SetText(amount2)
		},
	)

	table := widget.NewTable(
		func() (int, int) { return len(pairs), 7 },
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell 000, 000")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			pair := pairs[id.Row]
			trade := trades[pair]

			if trade.Data.Swaps != nil {
				// n := unitrade.Name(trade.Data.Swaps[0])
				p, _ := unitrade.Price(trade.Data.Swaps[0])
				_, c := unitrades.WholePriceChanges(trade)
				_, _, d := unitrades.Duration(trade)

				switch id.Col {
				case 0:
					label.SetText(fmt.Sprintf("%d", id.Row+1))
				case 1:
					label.SetText(pair)
				case 2:
					label.SetText(fmt.Sprintf("%f", p))
				case 3:
					label.SetText(fmt.Sprintf("%f", c))
				case 4:
					label.SetText(fmt.Sprintf("%f", d))
				default:
					label.SetText(fmt.Sprintf("Cell %d, %d", id.Row+1, id.Col+1))
				}
			}
		})
	table.SetColumnWidth(0, 34)
	table.SetColumnWidth(1, 202)

	// leftList.OnSelected = func(id widget.ListItemID) {
	// 	activePair, _ = pairsdata.GetValue(id)
	// 	selected = trades[activePair]
	// 	rightList.Refresh()
	// }

	listPanel := container.NewBorder(nil, control, nil, nil, table)

	var swaps uniswap.Swaps
	cc := make(chan string, 1)

	go func() {
		for _, pair := range pairs {
			go uniswap.SwapsByCounts(cc, 10, pair)

			msg := <-cc
			json.Unmarshal([]byte(msg), &swaps)

			if len(swaps.Data.Swaps) == 0 {
				time.Sleep(time.Second * 5)
				continue
			}
			if len(trades[pair].Data.Swaps) > 0 &&
				(trades[pair].Data.Swaps[0].Id == swaps.Data.Swaps[0].Id ||
					trades[pair].Data.Swaps[0].Timestamp > swaps.Data.Swaps[0].Timestamp) {
				continue
			}

			n := unitrade.Name(swaps.Data.Swaps[0])
			p, _ := unitrade.Price(swaps.Data.Swaps[0])
			_, c := unitrades.WholePriceChanges(swaps)
			_, _, d := unitrades.Duration(swaps)

			if activePair == pair &&
				selected.Data.Swaps[0].Id != swaps.Data.Swaps[0].Id {
				selected = swaps
				rightList.Refresh()
			}
			trades[pair] = swaps

			if p != oldPrices[pair] {
				t := time.Now()
				message := fmt.Sprintf("%s: %f %f %f", n, p, c, d)
				title := "Priced Up!"
				if c < 0 {
					title = "Priced Down!"
				}
				link := fmt.Sprintf("https://www.dextools.io/app/ether/pair-explorer/%s", pair)

				min, max := data.ReadMinMax(pair)

				if p < min {
					title = fmt.Sprintf("Warning Low! Watch %s", n)
				}
				if p > max {
					title = fmt.Sprintf("Warning High! Watch %s", n)
				}

				services.Alert(title, message, link, gosxnotifier.Morse)

				fmt.Println(".")
				fmt.Println(t.Format("2006/01/02 15:04:05"), ": ", n, p, c, d)
				fmt.Println(".")
			}
			oldPrices[pair] = p
			table.Refresh()
			time.Sleep(time.Second * 1)
		}
	}()

	return container.NewHSplit(listPanel, rightList)
}
