package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/template/html/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbString := os.Getenv("DB")
	if dbString == "" {
		log.Fatalln("Postgres connection string not found")
	}

	db, err := pgx.Connect(ctx, dbString)
	if err != nil {
		log.Fatalln(err.Error())
	}

	_, err = db.Exec(ctx, `
		create table if not exists items (id varchar(255), name varchar(255));
	`)

	if err != nil {
		log.Fatalln(err.Error())
	}

	engine := html.New("ui", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/healthcheck", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	app.Get("/", func(c fiber.Ctx) error {
		rows, err := db.Query(c.Context(), "select id, name from items;")
		if err != nil {
			log.Printf("Failed select items. Error: %v", err)
			return c.SendStatus(http.StatusInternalServerError)
		}
		defer rows.Close()

		items := make([]*entity, 0)

		for rows.Next() {
			item := new(entity)
			err = rows.Scan(
				&item.Id,
				&item.Name,
			)
			if err != nil {
				log.Printf("Failed scan item. Error: %v", err)
				return c.SendStatus(http.StatusInternalServerError)
			}
			items = append(items, item)
		}

		return c.Render("home", fiber.Map{
			"Items": items,
		})
	})

	app.Post("/", func(c fiber.Ctx) error {
		text := c.FormValue("text")
		if text != "" {
			_, err := db.Exec(c.Context(), "insert into items(id, name) values($1, $2)", uuid.New().String(), text)
			if err != nil {
				log.Printf("Failed create item. Error: %v", err)
				return c.SendStatus(http.StatusInternalServerError)
			}
		}

		itemId := c.FormValue("delete_index")
		if itemId != "" {
			_, err := db.Exec(c.Context(), "delete from items where id = $1", itemId)
			if err != nil {
				log.Printf("Failed delete item. Error: %v", err)
				return c.SendStatus(http.StatusInternalServerError)
			}
		}

		c.Set("Location", "/")
		return c.SendStatus(http.StatusFound)
	})

	log.Println("Start http server on port 8080")
	err = app.Listen(":8080", fiber.ListenConfig{})
	if err != nil {
		log.Fatalln(err.Error())
	}
}

type entity struct {
	Id   string
	Name string
}
