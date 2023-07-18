module github.com/Tixologi-Inc/gobrake/examples/negroni

go 1.15

replace github.com/Tixologi-Inc/gobrake => ../..

replace github.com/airbrake/gobrake/negroni => ../../negroni

require (
	github.com/airbrake/gobrake/negroni v0.0.0-00010101000000-000000000000
	github.com/Tixologi-Inc/gobrake v5.0.2
	github.com/gorilla/mux v1.8.0
	github.com/urfave/negroni v1.0.0
)
