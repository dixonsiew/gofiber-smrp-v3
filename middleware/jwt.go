package middleware

import (
	"smrp/model"
	tokenService "smrp/service/token"
	userService "smrp/service/user"
	"smrp/utils"
	"strings"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
)

func JWTProtected(c fiber.Ctx) error {
    authHeader := c.Get(fiber.HeaderAuthorization)
    if authHeader != "" {
        return jwtware.New(jwtware.Config{
            SigningKey: jwtware.SigningKey{Key: []byte(utils.JWT_SECRET)},
            //ContextKey: "jwt",
            Extractor: extractors.FromAuthHeader("Bearer"),
            ErrorHandler: func(c fiber.Ctx, err error) error {
                return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                    "statusCode": fiber.StatusUnauthorized,
                    "message":    "Unauthorized",
                })
            },
        })(c)
    }

    token := c.Cookies("token")
    if token == "" {
        authHeader := c.Get(fiber.HeaderAuthorization)
        if authHeader != "" {
            token = strings.ReplaceAll(authHeader, "Bearer ", "")
        }
    }

    if token == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "statusCode": fiber.StatusUnauthorized,
            "message":    "Unauthorized",
        })
    }

    userId, user, err := ValidateTokenStr(token)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "statusCode": fiber.StatusUnauthorized,
            "message":    "Invalid or expired token",
        })
    }

    c.Locals("userId", userId)
    c.Locals("username", user.Username)
    return c.Next()
}

func ValidateTokenStr(token string) (int64, *model.User, error) {
    _, id, err := tokenService.DecodeTokenStr(token)
    if err != nil {
        return id, nil, err
    }

    user, err := userService.FindById(id)
    if err != nil || user == nil {
        return id, user, err
    }

    return id, user, nil
}

func ValidateToken(c fiber.Ctx) (int64, *model.User, error) {
    _, id, err := tokenService.DecodeToken(c)
    if err != nil {
        return id, nil, err
    }

    user, err := userService.FindById(id)
    if err != nil || user == nil {
        return id, user, err
    }

    return id, user, nil
}

func NoContent(c fiber.Ctx) error {
    return c.Status(fiber.StatusNoContent).JSON(fiber.Map{
        "statusCode": fiber.StatusNoContent,
        "message":    "",
    })
}

func Unauthorized(c fiber.Ctx) error {
    mx := fiber.Map{
        "statusCode": fiber.StatusUnauthorized,
        "message":    "User Not Found",
    }
    return c.Status(fiber.StatusUnauthorized).JSON(mx)
}
