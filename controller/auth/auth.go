package auth

import (
    "smrp/dto"
    "smrp/middleware"
    tokenService "smrp/service/token"
    userService "smrp/service/user"
    "smrp/utils"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v3"
)

// Login
//
// @Tags Auth
// @Produce json
// @Param request body dto.LoginDto true "Login Request"
// @Success 200
// @Router /o/token [post]
func Login(c fiber.Ctx) error {
    data := new(dto.LoginDto)
    mx := fiber.Map{
        "statusCode": fiber.StatusUnauthorized,
        "message":    "Invalid Credentials",
    }
    if err := c.Bind().Body(data); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            errs := utils.GetValidationErrors(validationErrors)
            if errs != nil {
                return c.Status(fiber.StatusUnauthorized).JSON(mx)
            }
        }

        return c.Status(fiber.StatusUnauthorized).JSON(mx)
    }

    user, err := userService.FindByUsername(data.Username)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(mx)
    }

    valid := false
    if user != nil {
        valid = userService.ValidateCredentials(*user, data.Password)
    }

    if !valid {
        return c.Status(fiber.StatusUnauthorized).JSON(mx)
    }

    a := *user
    userService.UpdateLastLogin(a.Id.Int64)
    token, err := tokenService.GenerateAccessToken(a)
    refreshToken, errx := tokenService.GenerateRefreshToken(a)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(mx)
    }

    if errx != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(mx)
    }

    return c.JSON(fiber.Map{
        "type":          "bearer",
        "token":         token,
        "refresh_token": refreshToken,
    })
}

// Refresh
//
// @Tags Auth
// @Produce json
// @Param request body dto.RefreshTokenDto true "Refresh Token Request"
// @Security BearerAuth
// @Success 200
// @Router /o/refresh-token [post]
func Refresh(c fiber.Ctx) error {
    data := new(dto.RefreshTokenDto)
    mx := fiber.Map{
        "statusCode": fiber.StatusUnauthorized,
        "message":    "Invalid Credentials",
    }
    if err := c.Bind().Body(data); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            errs := utils.GetValidationErrors(validationErrors)
            if errs != nil {
                return c.Status(fiber.StatusUnauthorized).JSON(mx)
            }
        }

        return c.Status(fiber.StatusUnauthorized).JSON(mx)
    }

    md, err := tokenService.CreateAccessTokenFromRefreshToken(data.RefreshToken)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(mx)
    }

    return c.JSON(fiber.Map{
        "type":  "bearer",
        "token": md["token"],
    })
}

// UserDetails
//
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200
// @Router /api/current-user [get]
func UserDetails(c fiber.Ctx) error {
    _, user, err := middleware.ValidateToken(c)
    if err != nil {
        return err
    }

    if user == nil {
        return fiber.NewError(fiber.StatusNotFound, "User not found")
    }

    a := *user
    return c.JSON(fiber.Map{
        "id": a.Id,
        "username": a.Username,
        "first_name": a.Firstname,
        "last_name": a.Lastname,
        "roles": a.Roles,
    })
}

// ChangePassword
//
// @Tags Auth
// @Produce json
// @Param request body dto.ChangePasswordDto true "Change Password Request"
// @Security BearerAuth
// @Success 200
// @Router /api/change-password [post]
func ChangePassword(c fiber.Ctx) error {
    _, user, err := middleware.ValidateToken(c)
    if err != nil {
        return err
    }

    if user == nil {
        return fiber.NewError(fiber.StatusNotFound, "User not found")
    }

    data := new(dto.ChangePasswordDto)
    if err := c.Bind().Body(data); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            errs := utils.GetValidationErrors(validationErrors)
            if errs != nil {
                return fiber.NewError(fiber.StatusBadRequest, errs.Error())
            }
        }

        return fiber.NewError(fiber.StatusBadRequest, err.Error())
    }

    if data.Password != data.ConfirmPassword {
        return fiber.NewError(fiber.StatusBadRequest, "Confirm Password does not match")
    }

    a := *user
    a.Password = utils.NewNullString(data.Password)
    userService.UpdatePassword(a)
    return c.JSON(fiber.Map{
        "success": 1,
    })
}
