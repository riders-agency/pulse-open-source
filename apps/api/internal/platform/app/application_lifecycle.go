package app

func (a *Application) Close() {
	if a == nil || a.Deps.DB == nil {
		return
	}
	a.Deps.DB.Close()
}
