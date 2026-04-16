package user

import (
    "database/sql"
    "fmt"
    "smrp/database"
    "smrp/model"
    "smrp/utils"

    "golang.org/x/crypto/bcrypt"
)

func FindById(id int64) (*model.User, error) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil, nil
    }

    var o model.User
    q := `select id, username, first_name, last_name, password, last_login from app_user where id = $1 limit 1`
    err := db.GetContext(database.GetCtx(), &o, q, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, err
        }
        utils.LogError(err)
        return nil, err
    }

    o.Set(db)
    return &o, nil
}

func FindByUsername(username string) (*model.User, error) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil, nil
    }

    var o model.User
    q := `select id, username, first_name, last_name, password, last_login from app_user where username = $1 limit 1`
    err := db.GetContext(database.GetCtx(), &o, q, username)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, err
        }
        utils.LogError(err)
        return nil, err
    }

    o.Set(db)
    return &o, nil
}

func FindAll(offset int, limit int, sortBy string, sortDir string) ([]model.User, error) {
    lx := make([]model.User, 0)
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return lx, nil
    }

    q := `select t.id, t.username, t.first_name, t.last_name, t.password, t.last_login from app_user t order by %s %s offset $1 limit $2`
    q = fmt.Sprintf(q, sortBy, sortDir)
    err := db.SelectContext(database.GetCtx(),  &lx, q, offset, limit)
    if err != nil {
        utils.LogError(err)
        return nil, err
    }

    for i := range lx {
        lx[i].Set(db)
    }

    return lx, nil
}

func Count() (int, error) {
    n := 0
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return n, nil
    }

    q := `select count(id) from app_user`
    err := db.GetContext(database.GetCtx(), &n, q)
    if err != nil {
        utils.LogError(err)
        return n, err
    }

    return n, nil
}

func ExistsByOtherUsername(username string, id int64) (bool, error) {
    b := false
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return b, nil
    }

    q := `select exists (select 1 from app_user t where t.username = $1 and t.id <> $2)`
    err := db.QueryRowxContext(database.GetCtx(), q, username, id).Scan(&b)
    if err != nil {
        utils.LogError(err)
        return b, err
    }

    return b, nil
}

func ExistsByUsername(username string) (bool, error) {
    b := false
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return b, nil
    }

    q := `select exists (select 1 from app_user t where t.username = $1)`
    err := db.QueryRowxContext(database.GetCtx(), q, username).Scan(&b)
    if err != nil {
        utils.LogError(err)
        return b, err
    }

    return b, nil
}

func FindByKeyword(keyword string, offset int, limit int, sortBy string, sortDir string) ([]model.User, error) {
    lx := make([]model.User, 0)
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return lx, nil
    }

    q := `select t.id, t.username, t.first_name, t.last_name, t.password, t.last_login from app_user t where (t.username ilike $1 or t.first_name ilike $2 or t.last_name ilike $3) order by %s %s offset $4 limit $5`
    q = fmt.Sprintf(q, sortBy, sortDir)
    err := db.SelectContext(database.GetCtx(), &lx, q, keyword, keyword, keyword, offset, limit)
    if err != nil {
        utils.LogError(err)
        return lx, err
    }

    for i := range lx {
        lx[i].Set(db)
    }

    return lx, nil
}

func CountByKeyword(keyword string) (int, error) {
    n := 0
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return n, nil
    }

    q := `select count(id) from app_user t where (t.username ilike $1 or t.first_name ilike $2 or t.last_name ilike $3)`
    err := db.GetContext(database.GetCtx(), &n, q, keyword, keyword, keyword)
    if err != nil {
        utils.LogError(err)
        return n, err
    }

    return n, nil
}

func Save(o model.User) error {
    pw := []byte(o.Password.String)
    pwd, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
    if err != nil {
        utils.LogError(err)
        return err
    }

    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil
    }

    ctx := database.GetCtx()
    tx, err := db.BeginTxx(ctx, nil)
    if err != nil {
        utils.LogError(err)
        return err
    }

    defer func() {
        if err != nil {
            utils.LogError(err)
            tx.Rollback()
        }
    }()

    q := `insert into app_user (id, username, password, first_name, last_name, active) values(nextval('app_user_id_seq'),$1,$2,$3,$4,$5) returning id as app_user_id`
    var id int64
    err = tx.GetContext(ctx, &id, q, o.Username, pwd, o.Firstname, o.Lastname, true)
    if err != nil {
        utils.LogError(err)
        return err
    }

    for _, r := range o.Roles {
        qr := `insert into app_user_roles (app_user_id, roles_id) values($1, $2)`
        _, err := tx.ExecContext(ctx, qr, id, r.Id)
        if err != nil {
            utils.LogError(err)
            return err
        }
    }

    return tx.Commit()
}

func Update(o model.User) error {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil
    }

    ctx := database.GetCtx()
    tx, err := db.BeginTxx(ctx, nil)
    if err != nil {
        utils.LogError(err)
        return err
    }

    defer func() {
        if err != nil {
            utils.LogError(err)
            tx.Rollback()
        }
    }()

    if o.Password.Valid {
        pw := []byte(o.Password.String)
        pwd, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
        if err != nil {
            utils.LogError(err)
            return err
        }

        q := `update app_user set password = $1, first_name = $2, last_name = $3 where id = $4`
        _, err = tx.ExecContext(ctx, q, pwd, o.Firstname, o.Lastname, o.Id)
        if err != nil {
            utils.LogError(err)
            return err
        }
    } else {
        q := `update app_user set first_name = $1, last_name = $2 where id = $3`
        _, err = tx.ExecContext(ctx, q, o.Firstname, o.Lastname, o.Id)
        if err != nil {
            utils.LogError(err)
            return err
        }
    }
    
    q := `delete from app_user_roles where app_user_id = $1`
    _, err = tx.ExecContext(ctx, q, o.Id)
    if err != nil {
        utils.LogError(err)
        return err
    }

    for _, r := range o.Roles {
        qr := `insert into app_user_roles (app_user_id, roles_id) values($1,$2)`
        _, err := tx.ExecContext(ctx, qr, o.Id, r.Id)
        if err != nil {
            utils.LogError(err)
            return err
        }
    }

    return tx.Commit()
}

func DeleteById(id int64) error {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil
    }

    ctx := database.GetCtx()
    tx, err := db.BeginTxx(ctx, nil)
    if err != nil {
        utils.LogError(err)
        return err
    }

    defer func() {
        if err != nil {
            utils.LogError(err)
            tx.Rollback()
        }
    }()

    q := `delete from app_user_roles where app_user_id = $1`
    _, err = tx.ExecContext(ctx, q, id)
    if err != nil {
        utils.LogError(err)
        return err
    }

    q = `delete from app_user where id = $1`
    _, err = tx.ExecContext(ctx, q, id)
    if err != nil {
        utils.LogError(err)
        return err
    }

    return tx.Commit()
}

func UpdateLastLogin(id int64) error {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil
    }

    q := `update app_user set last_login = now() where id = $1`
    _, err := db.ExecContext(database.GetCtx(), q, id)
    if err != nil {
        utils.LogError(err)
        return err
    }

    return nil
}

func UpdatePassword(o model.User) error {
    pw := []byte(o.Password.String)
    pwd, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
    if err != nil {
        utils.LogError(err)
        return err
    }

    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return nil
    }

    q := `update app_user set password = $1 where id = $2`
    _, err = db.ExecContext(database.GetCtx(), q, pwd, o.Id)
    if err != nil {
        utils.LogError(err)
        return err
    }

    return nil
}

func ValidateCredentials(user model.User, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(password))
    return err == nil
}
