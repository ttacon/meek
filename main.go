package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nsf/termbox-go"
	"github.com/ttacon/innotop/innotop"
)

var (
	db *sql.DB

	// flags
	username = flag.String("u", "", "username")
	password = flag.String("p", "", "password")
)

func main() {
	flag.Parse()

	var err error
	if db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/", *username, *password)); err != nil {
		fmt.Println("failed to connect to db:", err)
		return
	}

	go monitorTxs(db)

	tx, err := db.Begin()
	if err != nil {
		fmt.Println("failed to begin transaction:", err)
		return
	}

	rows, err := tx.Query("select * from innotoptesting.hello")
	if err != nil {
		fmt.Println("failed to query table:", err)
	}

	<-time.After(time.Second * 15)

	rows.Close()
	tx.Rollback()
}

type columnInfo struct {
	name  string
	width int
}

var columns = []columnInfo{
	columnInfo{"ID", 20},
	columnInfo{"State", 20},
	columnInfo{"Started", 20},
	columnInfo{"MemoryUsed", 20},
}

func drawTitles(row int) {
	xCursor := 2
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
		xCursor += (5 + (col.width - colWidth))
	}

	for i := 0; i < xCursor; i++ {
		termbox.SetCell(i, row+1, '_', termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush()
}

func monitorTxs(db *sql.DB) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	drawTitles(1)
	count := 0

	for {
		<-time.After(time.Second)
		count++
		if count > 10 {
			break
		}
		rows, err := db.Query("select * from information_schema.INNODB_TRX")
		if err != nil {
			fmt.Println("failed to query transactions:", err)
			continue
		}

		// ID, State, Started, MemoryUsed
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

		}
		rows.Close()

	}
}
