package role

import (
	"database/sql"
	"fmt"
	"smrp/database"
	"smrp/model"
	"smrp/utils"
)

func FindAll(sortBy string, sortDir string) ([]model.Role, error) {
    lx := make([]model.Role, 0)
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return lx, nil
    }

    q := `select id, name from role order by %s %s`
    q = fmt.Sprintf(q, sortBy, sortDir)
    err := db.SelectContext(database.GetCtx(), &lx, q)
    if err != nil {
        utils.LogError(err)
        return nil, err
    }

    return lx, nil
}

func FindById(id int64) (*model.Role, error) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil, nil
    }

    var o model.Role
    q := `select id, name from role where id = $1 limit 1`
    err := db.GetContext(database.GetCtx(), &o, q, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, err
        }
        utils.LogError(err)
        return nil, err
    }

    return &o, nil
}
