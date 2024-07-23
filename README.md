# SurveyX

SurveyX is a web application for creating and managing surveys. This project uses Go and Gorilla Mux for the backend and PostgreSQL as the database.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Contributing](#contributing)
- [License](#license)

## Features

- Create and manage survey forms of  text,rating,mcq and checkbox types(also other types can be added)
- Add questions and options to surveys
- people can visit the surveys with live link and submit their responses
- analytics to anaylse the user responses and export cv option for storing data of responses in cv format
- users can make teams and add team members
- User authentication with Google OAuth
- Secure session management

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/nikhilsahni7/SurveyX-Backend.git
   cd surveyx
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Set up the database:

   Ensure you have PostgreSQL installed and running. Create a new database for the project.

4. Configure environment variables:

   Create a `.env` file in the root directory and add the following variables:

   ```properties
   export GOOGLE_CLIENT_ID="your-google-client-id"
   export GOOGLE_CLIENT_SECRET="your-google-client-secret"
   export OAUTH_STATE_STRING="your-oauth-state-string"
   export BASE_URL=http://localhost:8080
   export PORT=8080
   export DATABASE_URL="your-database-url"
   export SESSION_KEY="your-session-key"
   ```

5. Run the application:

   ```sh
   go run main.go
   ```

## Configuration

The application uses environment variables for configuration. These variables should be set in a `.env` file in the root directory of the project.

- `GOOGLE_CLIENT_ID`: Google OAuth client ID
- `GOOGLE_CLIENT_SECRET`: Google OAuth client secret
- `OAUTH_STATE_STRING`: A random string for OAuth state
- `BASE_URL`: The base URL of the application
- `PORT`: The port on which the application will run
- `DATABASE_URL`: The URL of the PostgreSQL database
- `SESSION_KEY`: A secret key for session management

## Usage

1. Start the server:

   ```sh
   go run main.go
   ```

2. Open your browser and navigate to `http://localhost:8080`.

3. Use the application to create and manage surveys.

## API Endpoints

- `POST /api/surveys`: Create a new survey
- `GET /api/surveys`: Get all surveys
- `GET /api/surveys/:id`: Get a specific survey by ID
- `PUT /api/surveys/:id`: Update a specific survey by ID
- `DELETE /api/surveys/:id`: Delete a specific survey by ID
- `POST /api/surveys/:id/duplicate`: Duplicate a specific survey by ID
- `POST /api/surveys/:id/publish`: Publish a specific survey by ID
- `POST /api/surveys/:id/unpublish`: Unpublish a specific survey by ID
- `POST /api/surveys/:id/submit`: Submit a response to a specific survey by ID
- `GET /api/surveys/:id/responses`: Get all responses for a specific survey by ID
- `GET /api/surveys/:id/responses/:responseId`: Get a specific response by response ID
- `GET /api/s/:linkID`: Access a survey by its public link ID
- `GET /api/surveys/:id/analytics`: Get analytics for a specific survey by ID
- `GET /api/surveys/:id/export`: Export survey data for a specific survey by ID
- `POST /api/teams`: Create a new team
- `GET /api/teams`: Get all teams
- `GET /api/teams/:teamId`: Get a specific team by ID
- `PUT /api/teams/:teamId`: Update a specific team by ID
- `POST /api/teams/:teamId/members`: Add a member to a specific team by ID
- `DELETE /api/teams/:teamId/members/:userId`: Remove a member from a specific team by user ID
- `POST /api/webhooks`: Create a new webhook
- `GET /api/webhooks`: Get all webhooks
- `PUT /api/webhooks/:id`: Update a specific webhook by ID
- `DELETE /api/webhooks/:id`: Delete a specific webhook by ID

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any changes.

1. Fork the repository
2. Create a new branch (`git checkout -b feature-branch`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to the branch (`git push origin feature-branch`)
5. Open a Pull Request

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
