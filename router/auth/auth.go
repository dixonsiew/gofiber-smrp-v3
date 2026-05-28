package auth

import (
    "smrp/controller/auth"
    "smrp/middleware"

    "github.com/gofiber/fiber/v3"
)

func SetupRoutes(router fiber.Router) {
    router.Post("/o/token", auth.Login)
    router.Post("/o/refresh-token", auth.Refresh)
    router.Get("/api/current-user", middleware.JWTProtected, auth.UserDetails)
    router.Post("/api/change-password", middleware.JWTProtected, auth.ChangePassword)
}
