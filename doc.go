// PGB: Pythia Gata Bot
// Copyright (C) 2019-2024  Yishen Miao
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

/*
Pgb is a telegram bot that generates random inline query results based on the user's query text and current time.
It only handles HTTP requests and should be run behind a reverse proxy that handes HTTPS termination.

This program expects the following environment variables:
	- HOST: The hostname or IP address where pgb binds to (default: "0.0.0.0").
	- PORT: The port on which pgb listens (default: "8080").
	- TOKEN: The Telegram bot authentication token (required).

Example usage:

Before running the application, ensure the required environment variables
are set:
	export HOST=127.0.0.1
	export PORT=8080
	export TOKEN=0123456789:abcdefghijklmnopqrstuvwxyz

Then run the application:
	go run pgb.go
*/

package main
