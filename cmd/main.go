package main

import (
	"order-payment-system/internal/app"
)

func main() {

	appl, port := app.InitializeApp()

	r := app.SetupRoutes(appl)

	app.Start(appl)

	r.Run(":" + port)

}
