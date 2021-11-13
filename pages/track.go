package pages

import (
	"encoding/json"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/uniswap-auto-gui/services"
	"github.com/uniswap-auto-gui/utils"
)

func trackScreen(_ fyne.Window) fyne.CanvasObject {
	dataList := binding.BindStringList(&[]string{"0x9d9681d71142049594020bd863d34d9f48d9df58", "0x7a99822968410431edd1ee75dab78866e31caf39"})

	append := widget.NewButton("Append", func() {
		dataList.Append("0x7a99822968410431edd1ee75dab78866e31caf39")
	})

	list := widget.NewListWithData(dataList,
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("address"), widget.NewLabel("token"), widget.NewLabel("price"), widget.NewLabel("change"), widget.NewLabel("duration"))
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			s := item.(binding.String)
			address := obj.(*fyne.Container).Objects[0].(*widget.Label)
			address.Bind(s)

			label := obj.(*fyne.Container).Objects[1].(*widget.Label)
			price := obj.(*fyne.Container).Objects[2].(*widget.Label)
			change := obj.(*fyne.Container).Objects[3].(*widget.Label)
			duration := obj.(*fyne.Container).Objects[4].(*widget.Label)

			var eth utils.Crypto
			var swaps utils.Swaps
			c1 := make(chan string)
			c2 := make(chan string)

			go func() {
				for {
					pair, _ := s.Get()
					utils.Post(c1, "bundles", "")
					utils.Post(c2, "swaps", pair)
					time.Sleep(time.Second * 2)
				}
			}()
			go func() {
				for {
					select {
					case msg1 := <-c1:
						json.Unmarshal([]byte(msg1), &eth)
						n, p, c, d, a := services.SwapsInfo(swaps)
						label.SetText(n)
						price.SetText(fmt.Sprintf("%f", p))
						change.SetText(fmt.Sprintf("%f", c))
						duration.SetText(fmt.Sprintf("%f hours", d))
						if a {
							services.Notify("Price Change Alert", n)
						}
					case msg2 := <-c2:
						json.Unmarshal([]byte(msg2), &swaps)
						n, p, c, d, a := services.SwapsInfo(swaps)
						label.SetText(n)
						price.SetText(fmt.Sprintf("%f", p))
						change.SetText(fmt.Sprintf("%f", c))
						duration.SetText(fmt.Sprintf("%f hours", d))
						if a {
							services.Notify("Price Change Alert", n)
						}
					}
				}
			}()
		})

	listPanel := container.NewBorder(nil, append, nil, nil, list)
	return container.NewGridWithColumns(1, listPanel)
}
