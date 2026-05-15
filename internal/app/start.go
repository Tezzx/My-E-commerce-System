package app

func Start(app *App) {
	app.orderTimeout.Start()
	app.orderCreate.Start()
}
