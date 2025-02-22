# How to contribute

## Set up development environment
1. Set up a local PostgreSQL database, i.e. using a container:
   ```shell
   docker run --name postgres \
		-v postgres:/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD=root \
		-p 5432:5432 \
		-d \
		postgres:16-alpine
   ```
1. Create databases named ```seatsurfing``` (for running the application) and ```seatsurfing_test``` (for running the tests) in your PostgreSQL database:
   ```sql
   CREATE database seatsurfing;
   CREATE database seatsurfing_test;
   ```
1. Check out Seatsurfing's code:
   ```shell
   git clone https://github.com/seatsurfing/seatsurfing.git
   cd seatsurfing
   ```
1. Typescript commons: Build the common typescript files:
   ```shell
   cd commons/ts && npm install && npm run build
   ```
1. Admin UI: Install dependencies and start the admin interface. Use a dedicated terminal for that:
   ```shell
   cd admin-ui
   npm install && npm run install-commons
   npm run dev
   ```
1. Booking UI: Install dependencies and start the booking interface. Use a dedicated terminal for that:
   ```shell
   cd booking-ui
   npm install && npm run install-commons
   npm run dev
   ```
1. Server: Install dependencies and run the server. Use a dedicated terminal for that:
   ```shell
   cd server
   go get .
   ./run.sh
   ```

You should now be able to access the Admin UI at http://localhost:8080/admin/ and the Booking UI at http://localhost:8080/ui/.
