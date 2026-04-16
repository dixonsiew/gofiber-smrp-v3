package user

import (
    "fmt"
    "smrp/dto"
    "smrp/model"
    roleService "smrp/service/role"
    userService "smrp/service/user"
    "smrp/utils"
    "strconv"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v3"
)

// List
//
// @Tags Setup/User
// @Produce json
// @Param        _page              query      string  false  "_page"
// @Param        _limit             query      string  false  "_limit"
// @Param        sort               query      string  false  "sort"
// @Security BearerAuth
// @Success 200 {array} model.User
// @Router /api/users [get]
func List(c fiber.Ctx) error {
    page := c.Query("_page", "1")
    limit := c.Query("_limit", "20")
    sorts := c.Query("sort", "")

    sortBy := "username"
    sortDir := "asc"

    if sorts != "" {
        lis := strings.Split(sorts, "$")
        s := lis[0]
        arr := strings.Split(s, ":")
        sortBy = arr[0]
        sortDir = arr[1]
    }

    total, err := userService.Count()
    if err != nil {
        return err
    }

    pg := model.GetPager(total, page, limit)
    lx, err := userService.FindAll(pg.GetLowerBound(), pg.PageSize, sortBy, sortDir)
    if err != nil {
        return err
    }

    c.Set(utils.X_TOTAL_COUNT, fmt.Sprintf("%d", total))
    c.Set(utils.X_TOTAL_PAGE, fmt.Sprintf("%d", pg.GetTotalPages()))
    return c.JSON(lx)
}

// SearchList
//
// @Tags Setup/User
// @Produce json
// @Param        _page    query      string         false  "_page"
// @Param        _limit   query      string         false  "_limit"
// @Param        sort     query      string         false  "sort"
// @Param        request  body       dto.KeywordDto true   "Search Request"
// @Security BearerAuth
// @Success 200 {array} model.User
// @Router /api/users [post]
func SearchList(c fiber.Ctx) error {
    page := c.Query("_page", "1")
    limit := c.Query("_limit", "20")
    sorts := c.Query("sort", "")

    data := new(dto.KeywordDto)
    if err := c.Bind().Body(data); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            errs := utils.GetValidationErrors(validationErrors)
            if errs != nil {
                return fiber.NewError(fiber.StatusBadRequest, err.Error())
            }
        }

        return fiber.NewError(fiber.StatusBadRequest, "Invalid Request")
    }

    key := fmt.Sprintf("%%%s%%", data.Keyword)
    sortBy := "username"
    sortDir := "asc"

    if sorts != "" {
        lis := strings.Split(sorts, "$")
        s := lis[0]
        arr := strings.Split(s, ":")
        sortBy = arr[0]
        sortDir = arr[1]
    }

    total, err := userService.CountByKeyword(key)
    if err != nil {
        return err
    }

    pg := model.GetPager(total, page, limit)
    lx, err := userService.FindByKeyword(key, pg.GetLowerBound(), pg.PageSize, sortBy, sortDir)
    if err != nil {
        return err
    }

    c.Set(utils.X_TOTAL_COUNT, fmt.Sprintf("%d", total))
    c.Set(utils.X_TOTAL_PAGE, fmt.Sprintf("%d", pg.GetTotalPages()))
    return c.JSON(lx)
}

// Create
//
// @Tags Setup/User
// @Produce json
// @Param        request  body       dto.UserDto true   "Create User Request"
// @Security BearerAuth
// @Success 200
// @Router /api/user [post]
func Create(c fiber.Ctx) error {
    data := new(dto.UserDto)
    if err := c.Bind().Body(data); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            errs := utils.GetValidationErrors(validationErrors)
            if errs != nil {
                return fiber.NewError(fiber.StatusBadRequest, errs.Error())
            }
        }

        return fiber.NewError(fiber.StatusBadRequest, err.Error())
    }

    username := data.Username
    b, err := userService.ExistsByUsername(username)
    if err != nil {
        return err
    }

    if b {
        return fiber.NewError(fiber.StatusBadRequest, "A user with that username already exists")
    }

    role, err := roleService.FindById(int64(data.RoleId))
    if err != nil {
        return err
    }

    if role == nil {
        return fiber.NewError(fiber.StatusNotFound, "Role not found")
    }

    o := model.User{
        Username:  utils.NewNullString(data.Username),
        Password:  utils.NewNullString(data.Password),
        Firstname: utils.NewNullString(data.Firstname),
        Lastname:  utils.NewNullString(data.Lastname),
        Roles:     []model.Role{*role},
    }
    userService.Save(o)
    return c.JSON(fiber.Map{
        "success": 1,
    })
}

// Edit
//
// @Tags Setup/User
// @Produce json
// @Param        id              path      int  true  "id"
// @Security BearerAuth
// @Success 200 {object} model.User
// @Router /api/user/{id} [get]
func Edit(c fiber.Ctx) error {
    ids := c.Params("id")
    id, _ := strconv.ParseInt(ids, 10, 64)
    o, err := userService.FindById(id)
    if err != nil {
        return err
    }

    if o == nil {
        return fiber.NewError(fiber.StatusNotFound, "User not found")
    }

    return c.JSON(o)
}

// Update
//
// @Tags Setup/User
// @Produce json
// @Param        id              path      int                true  "id"
// @Param        request         body      dto.UserDto true  "Update User Request"
// @Security BearerAuth
// @Success 200
// @Router /api/user/{id} [put]
func Update(c fiber.Ctx) error {
    ids := c.Params("id")
    id, _ := strconv.ParseInt(ids, 10, 64)
    data := new(dto.UserDto)
    if err := c.Bind().Body(data); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            errs := utils.GetValidationErrors(validationErrors)
            if errs != nil {
                return fiber.NewError(fiber.StatusBadRequest, errs.Error())
            }
        }

        return fiber.NewError(fiber.StatusBadRequest, err.Error())
    }

    username := data.Username
    password := data.Password
    b, err := userService.ExistsByOtherUsername(username, id)
    if err != nil {
        return err
    }

    if b {
        return fiber.NewError(fiber.StatusBadRequest, "A user with that username already exists")
    }

    role, err := roleService.FindById(int64(data.RoleId))
    if err != nil {
        return err
    }

    if role == nil {
        return fiber.NewError(fiber.StatusNotFound, "Role not found")
    }

    o := model.User{
        Id:        utils.NewInt64(id),
        Username:  utils.NewNullString(data.Username),
        Password:  utils.NewNullString(""),
        Firstname: utils.NewNullString(data.Firstname),
        Lastname:  utils.NewNullString(data.Lastname),
        Roles:     []model.Role{*role},
    }

    if password != "********" {
        o.Password = utils.NewNullString(password)
    }

    userService.Update(o)
    return c.JSON(fiber.Map{
        "success": 1,
    })
}

// Delete
//
// @Tags Setup/User
// @Produce json
// @Param        id              path      int                true  "id"
// @Security BearerAuth
// @Success 200
// @Router /api/user/{id} [delete]
func Delete(c fiber.Ctx) error {
    ids := c.Params("id")
    id, _ := strconv.ParseInt(ids, 10, 64)
    err := userService.DeleteById(id)
    if err != nil {
        return err
    }

    return c.JSON(fiber.Map{
        "success": 1,
    })
}
