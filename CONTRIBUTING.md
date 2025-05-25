# How to contribute

## Set up development environment

1. Make sure Node.js 22 or later is installed:

   ```shell
   node -v
   ```

1. Set up a local PostgreSQL database, i.e. using a container:

   ```shell
   docker run --name postgres \
   	-v postgres:/var/lib/postgresql/data \
   	-e POSTGRES_PASSWORD=root \
   	-p 5432:5432 \
   	-d \
   	postgres:16-alpine
   ```

1. Create databases named `seatsurfing` (for running the application) and `seatsurfing_test` (for running the tests) in your PostgreSQL database:

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

You should now be able to access the Admin UI at http://localhost:8080/admin/ and the Booking UI at http://localhost:8080/ui/. To login, use the default admin login (user `admin@seatsurfing.local` and password `12345678`).

## Running tests

If you add functionality (database queries, RESTful endpoints, utility functions etc.), please create corresponding unit tests - both positive and negative test cases.

If you modify existing backend functionality, please modify/add corresponding test cases.

If you add/modify major frontend functionality, please add/modify the e2e tests.

1. To run the backend/server unit tests:
   ```shell
   cd server
   ./test.sh
   ```

1. To run the e2e [Playwright](https://playwright.dev/) tests:
   1. Install the dependencies:
      ```shell
      cd e2e
      npm ci
      npx playwright install --with-deps
      ```

   1. Run the tests:
      ```shell
      npx playwright test
      ```

## Creating a pull request

Before submitting a pull request, please make sure the unit and e2e (written in [Playwright](https://playwright.dev/)) tests pass.

We use [conventional commits](https://www.conventionalcommits.org/) and squash merges, so the PR title should follow the conventional commit conventions.

Please provide a comprehensible description about the added/changed functionality. If frontend functionality is modified, screenshots are a welcome addition.