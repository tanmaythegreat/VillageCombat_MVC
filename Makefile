# 1. Define your giant URL as a reusable variable at the top
DB_URL=postgres://admin_tentellam:i_wont_tell_you@localhost:5432/VillageGameDB?sslmode=disable

# 2. The shortcut to push changes UP to the database
migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

# 3. The shortcut to roll changes DOWN (undo)
migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1