package commonsetup

import (
    "database/sql"
    "fmt"
    "smrp/database"
    "smrp/model"
    "smrp/utils"
)

func FindById(id int, table string) (*model.CommonSetup, error) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil, nil
    }

    var o model.CommonSetup
    q := `select id, code, "desc", ref, created_by, created_date, modified_by, modified_date, deleted, deleted_by, deleted_date from %s where id = $1 limit 1`
    q = fmt.Sprintf(q, table)
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

func FindByDesc(desc string, table string) (*model.CommonSetup, error) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil, nil
    }

    var o model.CommonSetup
    q := `select t.id, t.code, t.desc, t.ref, t.created_by, t.created_date, t.modified_by, t.modified_date, t.deleted, t.deleted_by, t.deleted_date from %s t where lower(t.desc) = lower($1) limit 1`
    q = fmt.Sprintf(q, table)
    err := db.GetContext(database.GetCtx(), &o, q, desc)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, err
        }
        utils.LogError(err)
        return nil, err
    }
    return &o, nil
}

func FindAll(table string, offset int, limit int, sortBy string, sortDir string) ([]model.CommonSetup, error) {
    lx := make([]model.CommonSetup, 0)
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return lx, nil
    }

    q := ""
    lp := make([]any, 0)
    if limit > 0 {
        q = `select t.id, t.code, t.desc, t.ref, t.created_by, t.created_date, t.modified_by, t.modified_date, t.deleted, t.deleted_by, t.deleted_date from %s t where t.deleted is not true order by "%s" %s offset $1 limit $2`
        q = fmt.Sprintf(q, table, sortBy, sortDir)
        lp = []any{ offset, limit }
    } else {
        q = `select t.id, t.code, t.desc, t.ref, t.created_by, t.created_date, t.modified_by, t.modified_date, t.deleted, t.deleted_by, t.deleted_date from %s t where t.desc <> '' and t.deleted is not true order by t.code`
        q = fmt.Sprintf(q, table)
    }

    err := db.SelectContext(database.GetCtx(), &lx, q, lp...)
    if err != nil {
        utils.LogError(err)
        return lx, err
    }

    return lx, nil
}

func Count(table string) (int, error) {
    n := 0
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return n, nil
    }

    q := `select count(id) from %s t where t.deleted is not true`
    q = fmt.Sprintf(q, table)
    err := db.GetContext(database.GetCtx(), &n, q)
    if err != nil {
        utils.LogError(err)
        return n, err
    }

    return n, nil
}

func FindByKeyword(keyword string, offset int, limit int, sortBy string, sortDir string, table string) ([]model.CommonSetup, error) {
    lx := make([]model.CommonSetup, 0)
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return lx, nil
    }

    q := `select t.id, t.code, t.desc, t.ref, t.created_by, t.created_date, t.modified_by, t.modified_date, t.deleted, t.deleted_by, t.deleted_date from %s t where (t.code ilike $1 or t.desc ilike $2 or t.ref ilike $3) and t.deleted is not true order by "%s" %s offset $4 limit $5`
    q = fmt.Sprintf(q, table, sortBy, sortDir)
    err := db.SelectContext(database.GetCtx(), &lx, q, keyword, keyword, keyword, offset, limit)
    if err != nil {
        utils.LogError(err)
        return lx, err
    }

    return lx, nil
}

func CountByKeyword(keyword string, table string) (int, error) {
    n := 0
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return n, nil
    }

    q := `select count(id) from %s t where (t.code ilike $1 or t.desc ilike $2 or t.ref ilike $3) and t.deleted is not true`
    q = fmt.Sprintf(q, table)
    err := db.GetContext(database.GetCtx(), &n, q, keyword, keyword, keyword)
    if err != nil {
        utils.LogError(err)
        return n, err
    }

    return n, nil
}

func Save(o model.CommonSetup, table string) error {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil
    }

    q := `insert into %s (id, code, "desc", ref, created_by, created_date) values(nextval('%s_id_seq'),$1,$2,$3,$4,now())`
    q = fmt.Sprintf(q, table, table)
    _, err := db.ExecContext(database.GetCtx(), q, o.Code.String, o.Desc.String, o.Ref.String, o.CreatedBy.Int64)
    if err != nil {
        utils.LogError(err)
        return err
    }

    return nil
}

func Update(o model.CommonSetup, table string) error {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil
    }

    q := `update %s set code = $1, "desc" = $2, ref = $3, modified_by = $4, modified_date = now() where id = $5`
    q = fmt.Sprintf(q, table)
    _, err := db.ExecContext(database.GetCtx(), q, o.Code.String, o.Desc.String, o.Ref.String, o.ModifiedBy.Int64, o.Id.Int64)
    if err != nil {
        utils.LogError(err)
        return err
    }

    return nil
}

func DeleteById(id int64, user_id int64, table string) error {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil
    }

    q := `update %s set deleted = true, deleted_by = $1, deleted_date = now() where id = $2`
    q = fmt.Sprintf(q, table)
    _, err := db.ExecContext(database.GetCtx(), q, user_id, id)
    if err != nil {
        utils.LogError(err)
        return err
    }

    return nil
}
