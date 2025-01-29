# Astungkara API

[![Test Coverage](https://img.shields.io/badge/coverage-100%25-brightgreen.svg)]()
[![Go Report Card](https://goreportcard.com/badge/github.com/nathakusuma/astungkara)]()

Astungkara API is a robust conference management system. The system provides a
comprehensive set of features for users, event coordinators, and administrators to manage conference sessions,
registrations, feedback, and more. The API is built using Go and Fiber, with PostgreSQL as the database and Redis for
temporary data storage.

## 🌟 Features

- Complete user authentication and profile management with OTP email verification
- Conference session management and registration
- Seat availability tracking
- Session proposal system with coordinator approval workflow
- Feedback management system
- Role-based access control (User, Event Coordinator, Admin)
- API documentation with Swagger
- System monitoring with Prometheus and Grafana
- High test coverage (100% for service layer)

## 🚀 Tech Stack

### Core

- **Language:** Go
- **Framework:** Fiber
- **Database:** PostgreSQL
- **Temporary Data Storage:** Redis
- **Web Server:** NGINX

### Libraries

- SQLX for database operations
- JWT for authentication
- go-playground/validator for input validation
- Zerolog for logging
- Viper for configuration management
- Testify for testing
- Gomail for email notifications

### DevOps

- Docker & Docker Compose
- Prometheus & Grafana for monitoring
- GNU Make for build automation

## 🛠 Installation

### Prerequisites

- Internet connection
- Docker and Docker Compose
- GNU Make
- Git

### Setup Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/nathakusuma/astungkara.git
   cd astungkara
   git checkout i-putu-natha-kusuma
   ```

2. Configure environment variables:
   ```bash
   cp .env.example .env
   # Edit .env file with your configuration
   ```

3. Start the application:
   ```bash
   make run
   ```

4. (Optional) Seed the database with test data:
   ```bash
   make seed-up dev
   ```

## 📚 Documentation

- API Documentation: `https://[HOSTNAME]/swagger`
- Monitoring Dashboard: `https://[HOSTNAME]/grafana`

## 🌐 Deployment

The API is deployed and accessible at:

```
https://astungkara.nathakusuma.com
```

## 🔒 Security

- Email verification with OTP
- JWT-based authentication
- Access & refresh token system
- Role-based access control
- Input validation
- Secure password hashing
- Rate limiting
- HTTPS enforcement

## 🔍 Monitoring

The system includes comprehensive monitoring:

- Real-time metrics with Prometheus
- Visualized dashboards with Grafana
- Performance monitoring
- Error tracking
- Resource utilization metrics

## 👥 Authors

- I Putu Natha Kusuma

---

## **⭐** Minimum Viable Product (MVP)

As we have mentioned earlier, we need technology that can support Conference in the future. Please consider these features below:

- New user can register account to the system ✔️
- User can login to the system ✔️
- User can edit their profile account ✔️
- User can view all conference sessions ✔️
- User can leave feedback on sessions ✔️
- User can view other user's profile ✔️
- Users can register for sessions during the conference registration period if seats are available ✔️
- Users can only register for one session within a time period ✔️
- Users can create, edit, delete their session proposals ✔️
- Users can only create one session proposal within a time period ✔️
- Users can edit, delete their session ✔️
- Event Coordinator can view all session proposals ✔️
- Event Coordinator can accept or reject user session proposals ✔️
- Event Coordinator can remove sessions ✔️
- Event Coordinator can remove user feedback ✔️
- Admin can add new event coordinators ✔️
- Admin can remove users/event coordinators ✔️

## **🌎** Service Implementation

```
GIVEN => I am a new user
WHEN  => I register to the system
THEN  => System will record and return the user's registration details

GIVEN => I am a user
WHEN  => I log in to the system
THEN  => System will authenticate and grant access based on user credentials

GIVEN => I am a user
WHEN  => I edit my profile account
THEN  => The system will update my account with the new information

GIVEN => I am a user
WHEN  => I view all available conference's sessions
THEN  => System will display all conference sessions with their details

GIVEN => I am a user
WHEN  => I leave feedback on a session
THEN  => System will record my feedback and display it under the session

GIVEN => I am a user
WHEN  => I view other user's profiles
THEN  => System will show the information of other user's profiles

GIVEN => I am a user
WHEN  => I register for conference's sessions
THEN  => System will confirm my registration for the selected session

GIVEN => I am a user
WHEN  => I create a new session proposal
THEN  => System will record and confirm the session creation

GIVEN => I am a user
WHEN => I see my session's proposal details
THEN => System will display my session's proposal details

GIVEN => I am a user
WHEN  => I update my session's proposal details
THEN  => System will apply the changes and confirm the update

GIVEN => I am a user
WHEN  => I delete my session's proposal
THEN  => System will remove the session's proposal

GIVEN => I am a user
WHEN => I see my session details
THEN => System will display my session details

GIVEN => I am a user
WHEN  => I update my session details
THEN  => System will apply the changes and confirm the update

GIVEN => I am a user
WHEN  => I delete my session
THEN  => System will remove the session

GIVEN => I am an event coordinator
WHEN  => I view session proposals
THEN  => System will display all submitted session proposals

GIVEN => I am an event coordinator
WHEN  => I accept or reject the session proposal
THEN  => The system will make the session either be available or unavailable

GIVEN => I am an event coordinator
WHEN  => I remove a session
THEN  => System will delete the session

GIVEN => I am an event coordinator
WHEN  => I remove user feedback
THEN  => System will delete the feedback from the session

GIVEN => I am an admin
WHEN  => I add new event coordinator
THEN  => System will make the account to the system

GIVEN => I am an admin
WHEN  => I remove a user or event coordinator
THEN  => System will delete the account from the system
```
