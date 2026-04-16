package utils

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "reflect"
    "regexp"
    "smrp/controller/report"
    "strconv"
    "strings"
    "time"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v3"
    "github.com/guregu/null/v6"
    "github.com/nleeper/goment"
    "github.com/rs/zerolog"
    "github.com/xuri/excelize/v2"
    "github.com/ztrue/tracerr"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

type StructValidator struct {
    Xvalidate *validator.Validate
}

func (v *StructValidator) Validate(out any) error {
    return v.Xvalidate.Struct(out)
}

var (
    Logger  zerolog.Logger
    iLogger zerolog.Logger
)

func SetLogger(runLogFile *os.File) {
    multi := zerolog.MultiLevelWriter(os.Stdout, runLogFile)
    Logger = zerolog.New(multi).Level(zerolog.ErrorLevel).With().Timestamp().Caller().Logger()

    iLogger = zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger()
}

func GetValidationErrors(errs validator.ValidationErrors) error {
    if len(errs) > 0 {
        errMsgs := make([]string, 0)
        for _, err := range errs {
            switch err.Tag() {
            case "required":
                ex := fmt.Sprintf("[%s] is %s", err.Field(), err.Tag())
                errMsgs = append(errMsgs, ex)
            case "max":
                ex := fmt.Sprintf("[%s] max length is %s", err.Field(), err.Param())
                errMsgs = append(errMsgs, ex)
            case "min":
                ex := fmt.Sprintf("[%s] min length is %s", err.Field(), err.Param())
                errMsgs = append(errMsgs, ex)
            default:
                errMsgs = append(errMsgs, fmt.Sprintf(
                    "[%s]: '%v' | Needs to implement '%s' '%s'",
                    err.Field(),
                    err.Value(),
                    err.Tag(),
                    err.Param(),
                ))
            }
        }

        return &fiber.Error{
            Code:    fiber.ErrBadRequest.Code,
            Message: strings.Join(errMsgs, " and "),
        }
    }

    return nil
}

func GetErrors(errs []error) string {
    ls := []string{}
    for _, err := range errs {
        ls = append(ls, err.Error())
    }

    return strings.Join(ls, "|")
}

func GetDateStr(v any) string {
    o := GetStr(v)
    k := reflect.TypeOf(v)
    if k.String() == "bson.DateTime" {
        iv, _ := strconv.Atoi(o)
        t := time.UnixMilli(int64(iv))
        g, _ := goment.New(t)
        gs := g.Format("YYYY-MM-DD")
        o = gs
    }

    s := o
    if len(o) >= 10 {
        s = o[0:10]
    }

    return s
}

func GetDateTimeStr(s string) string {
    return strings.ReplaceAll(s, "Z", "")
}

func NewNullString(s string) null.String {
    if s == "" {
        return null.NewString(s, false)
    }
    return null.NewString(s, true)
}

func NewInt32(i int32) null.Int32 {
    return null.NewInt32(i, true)
}

func NewInt64(i int64) null.Int64 {
    return null.NewInt(i, true)
}

func NewFloat(f float64) null.Float {
    return null.NewFloat(f, true)
}

func SetValue(x bson.M, ofield string, srcField string) {
    v := x[srcField].(string)
    _, ok := x[ofield]
    if ok {
        s := x[ofield].(string)
        if "N/A" == s {
            x[ofield] = v
        }
    } else {
        x[ofield] = v
    }

    if x[ofield] == "undefined" {
        x[ofield] = "N/A"
    }
}

func GetStr(a any) string {
    return fmt.Sprintf("%v", a)
}

func GetNumber(s string) int64 {
    i, _ := strconv.ParseInt(s, 10, 64)
    return i
}

func GetNum(s string) float64 {
    re := regexp.MustCompile(`[^\d.]*`)
    r := re.ReplaceAllString(s, "")
    v, _ := strconv.ParseFloat(r, 64)
    return v
}

func GetXlsx(colmaps []report.ColumnMap, lx []bson.M) (*bytes.Buffer, error) {
    f := excelize.NewFile()
    defer func() {
        if err := f.Close(); err != nil {
            LogError(err)
        }
    }()

    style, _ := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{
            Bold: true,
        },
    })
    coloffset := 6
    sh := "Sheet1"

    for i, cx := range colmaps {
        j := i + 1
        cellName, _ := excelize.CoordinatesToCellName(j, 1)
        f.SetCellStr(sh, cellName, cx.Text)
        f.SetCellStyle(sh, cellName, cellName, style)
        n := len(cx.Text)
        cola, _ := excelize.ColumnNumberToName(j)
        f.SetColWidth("Sheet1", cola, cola, float64(n+coloffset))
    }

    k := 2
    for _, x := range lx {
        for i, cx := range colmaps {
            field := cx.Field
            j := i + 1
            _, ok := x[field]
            s := ""
            if ok {
                s = GetStr(x[field])
            }

            cellName, _ := excelize.CoordinatesToCellName(j, k)
            f.SetCellStr(sh, cellName, s)
            n := len(s)
            cola, _ := excelize.ColumnNumberToName(j)
            m, _ := f.GetColWidth(sh, cola)
            if float64(n) > (m - float64(coloffset)) {
                f.SetColWidth("Sheet1", cola, cola, float64(n+coloffset))
            }
        }

        k = k + 1
    }

    bx, err := f.WriteToBuffer()
    return bx, err
}

func GetCollectionList(ctx context.Context, col *mongo.Collection, opts ...options.Lister[options.FindOptions]) ([]bson.M, error) {
    var ls []bson.M = make([]bson.M, 0)
    res, err := FindList(ctx, col, opts...)
    if err != nil {
        return ls, err
    }

    ls = processDoc(res)
    return ls, nil
}

func FindList(ctx context.Context, col *mongo.Collection, opts ...options.Lister[options.FindOptions]) ([]bson.M, error) {
    var res []bson.M = make([]bson.M, 0)
    var err error
    cur, err := col.Find(ctx, bson.D{}, opts...)
    if err != nil {
        LogError(err)
        return res, err
    }

    defer cur.Close(ctx)

    for cur.Next(ctx) {
        var doc bson.M
        err = cur.Decode(&doc)
        if err != nil {
            LogError(err)
            break
        }

        res = append(res, doc)
    }

    if err = cur.Err(); err != nil {
        LogError(err)
        return res, err
    }

    return res, err
}

func GetCollectionListWithFilter(ctx context.Context, filter any, col *mongo.Collection, opts ...options.Lister[options.FindOptions]) ([]bson.M, error) {
    var ls []bson.M = make([]bson.M, 0)
    var res []bson.M = make([]bson.M, 0)
    var err error
    cur, err := col.Find(ctx, filter, opts...)
    if err != nil {
        LogError(err)
        return res, err
    }

    defer cur.Close(ctx)

    for cur.Next(ctx) {
        var doc bson.M
        err = cur.Decode(&doc)
        if err != nil {
            LogError(err)
            break
        }

        res = append(res, doc)
    }

    if err = cur.Err(); err != nil {
        LogError(err)
        return res, err
    }

    ls = processDoc(res)
    return ls, err
}

func processDoc(lx []bson.M) []bson.M {
    ls := make([]bson.M, 0)
    na := "N/A"
    for _, x := range lx {
        v, ok := x["ADMISSION_DATE"]
        if ok {
            x["ADMISSION_DATE"] = GetDateStr(v)
        }

        v, ok = x["DISCHARGE_DATE"]
        if ok {
            x["DISCHARGE_DATE"] = GetDateStr(v)
        }

        v, ok = x["DEATH_DATE"]
        if ok {
            x["DEATH_DATE"] = GetDateStr(v)
        }

        v, ok = x["DELIVERY_DATE"]
        if ok {
            x["DELIVERY_DATE"] = GetDateStr(v)
        }

        v, ok = x["PATIENT_NOK_NAME"]
        if ok {
            s := v.(string)
            if na == s {
                x["NOK_STREET1"] = na
                x["NOK_STREET2"] = na
                x["NOK_CITYCODE"] = na
                x["NOK_POSTCODE"] = na
                x["NOK_OCITY"] = na
                x["NOK_NATIONALITY"] = na
            } else {
                SetValue(x, "NOK_STREET1", "STREET1")
                SetValue(x, "NOK_STREET2", "STREET2")
                SetValue(x, "NOK_CITYCODE", "CITYCODE")
                SetValue(x, "NOK_POSTCODE", "POSTCODE")
                SetValue(x, "NOK_OCITY", "OCITY")
                SetValue(x, "NOK_NATIONALITY", "NATIONALITY")
            }
        } else {
            x["NOK_STREET1"] = na
            x["NOK_STREET2"] = na
            x["NOK_CITYCODE"] = na
            x["NOK_POSTCODE"] = na
            x["NOK_OCITY"] = na
            x["NOK_NATIONALITY"] = na
        }

        ls = append(ls, x)
    }

    return ls
}

func CatchPanic(funcName string) {
    if err := recover(); err != nil {
        LogError(fmt.Errorf("recovered from panic -%s:%v", funcName, err))
    }
}

func LogError(err error) {
    ex := tracerr.Wrap(err)
    Logger.Err(err).Msg(tracerr.Sprint(ex))
}

func LogInfo(s string) {
    iLogger.Info().Msg(s)
}
