package main

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    // echo instance
    router := echo.New()

    // middleware
    // router.Use( middleware.Logger() ) // This is verbose but reveals useful information
    router.Use( middleware.Recover() )

    registerMicroserviceFunctions( router ) ;

    // start server
    router.Logger.Fatal( router.Start(":80") )
}