package model

import (
    "math"
    "smrp/database"
    "smrp/utils"
    "strconv"

    "github.com/guregu/null/v6"
    "github.com/jmoiron/sqlx"
)

type Pager struct {
    Total    int
    PageNum  int
    PageSize int
}

func (o *Pager) SetPageSize(pageSize int) {
    if (o.Total < pageSize || pageSize < 1) && o.Total > 0 {
        o.PageSize = o.Total
    } else {
        o.PageSize = pageSize
    }

    if o.GetTotalPages() < o.PageNum {
        o.PageNum = o.GetTotalPages()
    }

    if o.PageNum < 1 {
        o.PageNum = 1
    }
}

func (o *Pager) GetLowerBound() int {
    return (o.PageNum - 1) * o.PageSize
}

func (o *Pager) GetUpperBound() int {
    x := o.PageNum * o.PageSize
    if o.Total < x {
        x = o.Total
    }

    return x
}

func (o *Pager) GetTotalPages() int {
    v := float64(o.Total) / float64(o.PageSize)
    x := math.Ceil(v)
    return int(x)
}

func GetPager(total int, page string, limit string) Pager {
    pageNum, _ := strconv.Atoi(page)
    pageSize, _ := strconv.Atoi(limit)
    pg := Pager{
        Total:    total,
        PageNum:  pageNum,
        PageSize: pageSize,
    }
    pg.SetPageSize(pageSize)
    return pg
}

type Role struct {
    Id   null.Int64  `json:"id" db:"id" swaggertype:"integer"`
    Name null.String `json:"name" db:"name" swaggertype:"string"`
}

type User struct {
    Id        null.Int64  `json:"id" db:"id" swaggertype:"integer"`
    Username  null.String `json:"username" db:"username" swaggertype:"string"`
    Firstname null.String `json:"first_name" db:"first_name" swaggertype:"string"`
    Lastname  null.String `json:"last_name" db:"last_name" swaggertype:"string"`
    Password  null.String `json:"-" db:"password" swaggertype:"string"`
    LastLogin null.String `json:"last_login" db:"last_login" swaggertype:"string"`
    Roles     []Role      `json:"roles"`
}

func (o *User) Set(db *sqlx.DB) {
    o.LastLogin = utils.NewNullString(utils.GetDateTimeStr(o.LastLogin.String))
    o.SetRoles(db)
}

func (o *User) SetRoles(db *sqlx.DB) {
    lx := make([]Role, 0)
    rows, _ := db.QueryxContext(database.GetCtx(), `select aur.app_user_id, aur.roles_id, r.id, r.name from app_user_roles aur inner join role r on aur.roles_id = r.id where aur.app_user_id = $1`, o.Id)
    defer rows.Close()
    for rows.Next() {
        var (
            app_user_id null.Int64
            roles_id    null.Int64
            id          null.Int64
            name        null.String
        )
        _ = rows.Scan(&app_user_id, &roles_id, &id, &name)
        k := Role{
            Id:   id,
            Name: name,
        }
        lx = append(lx, k)
    }

    o.Roles = lx
}

type CommonSetup struct {
    Id           null.Int64  `json:"id" db:"id" swaggertype:"integer"`
    Code         null.String `json:"code" db:"code" swaggertype:"string"`
    Desc         null.String `json:"desc" db:"desc" swaggertype:"string"`
    Ref          null.String `json:"ref" db:"ref" swaggertype:"string"`
    CreatedBy    null.Int64  `json:"created_by" db:"created_by" swaggertype:"integer"`
    CreatedDate  null.String `json:"created_date" db:"created_date" swaggertype:"string"`
    ModifiedBy   null.Int64  `json:"modified_by" db:"modified_by" swaggertype:"integer"`
    ModifiedDate null.String `json:"modified_date" db:"modified_date" swaggertype:"string"`
    Deleted      null.Bool   `json:"deleted" db:"deleted" swaggertype:"boolean"`
    DeletedBy    null.Int64  `json:"deleted_by" db:"deleted_by" swaggertype:"integer"`
    DeletedDate  null.String `json:"deleted_date" db:"deleted_date" swaggertype:"string"`
}
