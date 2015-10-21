package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nsf/termbox-go"
	"github.com/ttacon/innotop"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	db *sql.DB

	// flags
	username          = flag.String("u", "", "username")
	host              = flag.String("h", "localhost", "host to connect to")
	password          = flag.String("-password", "", "password")
	port              = flag.Int("P", 3306, "port database server is bound to")
	promptForPassword = flag.Bool("p", false, "prompt for password")
)

func main() {
	flag.Parse()

	passwd := *password
	if *promptForPassword {
		fmt.Print("password:")
		passwdBytes, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println("failed to read password:", err)
			os.Exit(1)
		}
		passwd = string(passwdBytes)
	}

	var err error
	if db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", *username, passwd, *host, *port)); err != nil {
		fmt.Println("failed to connect to db:", err)
		os.Exit(1)
	} else if err = db.Ping(); err != nil {
		fmt.Println("failed to connect to db:", err)
		os.Exit(1)
	}

	events := make(chan keyEvent)
	if err = termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	go keyListener(events)
	monitor(db, events)
}

type mon func(*sql.DB, chan keyEvent) mon

func monitor(db *sql.DB, events chan keyEvent) {
	var todo = monitorTxs

	for {
		if todo = todo(db, events); todo == nil {
			break
		}
	}
}

type columnInfo struct {
	name  string
	width int
}

// ID, State, Started, Mem, Op State, RowsLocked, ROnly, Query
var txColumns = []columnInfo{
	columnInfo{"ID", 10},
	columnInfo{"State", 14},
	columnInfo{"Started", 20},
	columnInfo{"Mem", 11},
	columnInfo{"Op State", 10},
	columnInfo{"RowsLocked", 11},
	columnInfo{"ROnly", 6},
	columnInfo{"Query", 15},
}

// CXN        Cmd    ID         User  Host      DB   Time   Query
var qList = []columnInfo{
	columnInfo{"ID", 11},
	columnInfo{"Cmd", 7},
	columnInfo{"CXN", 11},
	columnInfo{"User", 8},
	columnInfo{"Host", 16},
	columnInfo{"DB", 8},
	columnInfo{"Time", 8},
	columnInfo{"Query", 30},
}

func drawTitles(typ keyEvent, row int) {
	xCursor := 2
	columns := txColumns

	w, _ := termbox.Size()
	for i := 0; i < w; i++ {
		termbox.SetCell(i, 1, ' ', colWhite, colDef)
	}

	if typ == queryList {
		columns = qList
	}

	for _, col := range columns {
		colWidth := 0
		for _, run := range col.name {
			termbox.SetCell(
				xCursor,
				row,
				run,
				termbox.ColorWhite,
				termbox.ColorDefault,
			)
			xCursor++
			colWidth++
		}

		// after every column add padding
		xCursor += (1 + (col.width - colWidth))
	}

	for i := 0; i < xCursor; i++ {
		termbox.SetCell(i, row+1, '_', termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush()
}

func keyListener(events chan keyEvent) {
	for {
		event := termbox.PollEvent()
		if event.Key == termbox.KeyCtrlC {
			go func() {
				events <- quitEvent
			}()
		} else if event.Type == termbox.EventKey {
			if event.Ch == 'Q' {
				go func() {
					events <- queryList
				}()
			} else if event.Ch == 'T' {
				go func() {
					events <- txList
				}()
			}
		}
	}
}

type keyEvent int

const (
	quitEvent keyEvent = iota
	queryList
	txList
)

func monitorTxs(db *sql.DB, events chan keyEvent) mon {
	drawTitles(txList, 1)

	oldRowNum := 0
	rowStart := 4 // incase we want to move stuff later?
	ww, _ := termbox.Size()

	ticker := time.Tick(time.Second)

	for {
		select {
		case eve := <-events:
			if eve == quitEvent {
				return nil
			} else if eve == queryList {
				clearRows(rowStart, rowStart+oldRowNum, ww)
				return monitorQueries
			}
		case <-ticker:

		}

		rows, err := db.Query("select * from information_schema.INNODB_TRX")
		if err != nil {
			fmt.Println("failed to query transactions:", err)
			continue
		}

		rowNum := rowStart

		clearRows(rowStart, rowStart+oldRowNum, ww)

		// ID, State, Started, MemoryUsed, Op State, RowsLocked, ROnly, Query
		for rows.Next() {
			var i innotop.InnoDBTransaction
			if err = rows.Scan(
				&i.ID, &i.State, &i.Started, &i.RequestedLockID,
				&i.WaitStarted, &i.Weight, &i.MySQLThreadID, &i.Query,
				&i.OperationState, &i.TablesInUse, &i.TablesLocked, &i.LockStructs,
				&i.LockMemoryBytes, &i.RowsLocked, &i.RowsModified, &i.ConcurrencyTickets,
				&i.IsolationLevel, &i.UniqueChecks, &i.ForeignKeyChecks, &i.LastForeignKeyError,
				&i.AdaptiveHashLatched, &i.AdaptiveHashTimeout, &i.IsReadOnly, &i.AutocommitNonLocking,
			); err != nil {
				fmt.Println("failed to read transaction:", err)
				break
			}

			// add the Transaction ID
			cursor := 2
			for _, run := range i.ID {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++
				if cursor == 10 {
					break
				}
			}

			cursor = 13
			for _, run := range i.State {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++
				if cursor == 26 {
					break
				}
			}

			cursor = 28
			for _, run := range string(i.Started) {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++
				if cursor == 49 {
					break
				}
			}

			cursor = 49
			for _, run := range strconv.FormatUint(i.LockMemoryBytes, 10) {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++
				if cursor == 59 {
					break
				}
			}

			cursor = 61
			for _, run := range i.OperationState.String {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++
				if cursor == 70 {
					break
				}
			}

			cursor = 72
			for _, run := range strconv.FormatUint(i.RowsLocked, 10) {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++
				if cursor == 76 {
					break
				}
			}

			cursor = 84
			isReadOnly := 'Y'
			if i.IsReadOnly == 0 {
				isReadOnly = 'N'
			}
			termbox.SetCell(cursor, rowNum, isReadOnly, colWhite, colDef)

			cursor = 91
			for _, run := range i.Query.String {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++
				if cursor == ww-1 {
					break
				}
			}

			rowNum++
		}
		oldRowNum = rowNum - rowStart
		rows.Close()
		termbox.Flush()
	}

	return nil
}

func clearRows(i0, iN, width int) {
	for i := i0; i < iN; i++ {
		for x := 0; x < width; x++ {
			termbox.SetCell(x, i, ' ', colWhite, colDef)
		}
	}
}

// term colors
var (
	colWhite = termbox.ColorWhite
	colDef   = termbox.ColorDefault
)

func monitorQueries(db *sql.DB, events chan keyEvent) mon {
	drawTitles(queryList, 1)

	oldRowNum := 0
	rowStart := 4 // incase we want to move stuff later?
	ww, _ := termbox.Size()
	ticker := time.Tick(time.Second)

	for {
		select {
		case eve := <-events:
			if eve == quitEvent {
				return nil
			} else if eve == txList {
				clearRows(rowStart, rowStart+oldRowNum, ww)
				return monitorTxs
			}
		case <-ticker:

		}

		rows, err := db.Query("select * from information_schema.PROCESSLIST")
		if err != nil {
			fmt.Println("failed to query transactions:", err)
			continue
		}

		rowNum := rowStart

		// clear old rows
		clearRows(rowStart, rowStart+oldRowNum, ww)

		// ID, State, Started, MemoryUsed
		for rows.Next() {
			var i innotop.ProcessInfo
			if err = rows.Scan(
				&i.ID, &i.User, &i.Host, &i.DB,
				&i.Command, &i.Time, &i.State, &i.Info,
				&i.TimeMS, &i.Stage, &i.MaxStage, &i.Progress,
				&i.MemoryUsed, &i.ExaminedRow, &i.QueryID,
			); err != nil {
				fmt.Println("failed to read transaction:", err)
				break
			}

			if i.Info.String == "select * from information_schema.PROCESSLIST" {
				continue
			}

			// add the Transaction ID
			cursor := 2
			for _, run := range strconv.FormatInt(i.QueryID, 10) {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++

				if cursor == 12 {
					break
				}
			}

			cursor = 14
			for _, run := range i.Command {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++

				if cursor == 20 {
					break
				}
			}

			cursor = 22
			for _, run := range i.State.String {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++

				if cursor == 32 {
					break
				}
			}

			cursor = 34
			for _, run := range i.User {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++

				if cursor == 31 {
					break
				}
			}

			cursor = 43
			for _, run := range i.Host {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++

				if cursor == 58 {
					break
				}
			}

			cursor = 60
			for _, run := range i.DB.String {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++

				if cursor == 67 {
					break
				}
			}

			cursor = 69
			for _, run := range tim(i.Time) {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++

				if cursor == 77 {
					break
				}
			}

			cursor = 79
			for _, run := range i.Info.String {
				termbox.SetCell(cursor, rowNum, run, colWhite, colDef)
				cursor++
			}

			rowNum++
		}
		oldRowNum = rowNum - rowStart
		rows.Close()
		termbox.Flush()
	}

	return nil
}

func tim(sec32 int32) string {
	sec := int(sec32)
	var seconds, mins, hours int
	if sec > 0 {
		seconds = sec % 60
		sec = sec / 60
	}

	if sec > 0 {
		mins = sec % 60
		sec = sec / 60
	}

	if sec > 0 {
		hours = sec % 60
	}

	return fmt.Sprintf("%0.2d:%0.2d:%0.2d", hours, mins, seconds)
}
