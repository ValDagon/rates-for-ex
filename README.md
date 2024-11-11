# Real-Time Currency Rates

A Go application that displays real-time currency exchange rates for **BTC to USD** and **USD to EUR**. The rates are updated every 10 seconds and presented on interactive charts.

## Project Description

This web application fetches current exchange rates from public APIs and displays them on interactive charts using Plotly.js. The charts are arranged horizontally:

- **Left**: BTC to USD exchange rate chart.
- **Right**: USD to EUR exchange rate chart.

### Key Features

- **Real-Time Updates**: Fetches new data every 10 seconds.
- **Interactive Charts**: Users can zoom, pan, and hover over data points to see detailed information.
- **Intuitive Interface**: Horizontally arranged charts for easy comparison.
- **Robustness**: Handles errors gracefully and processes system signals for stable operation.
- **Clean Architecture**: Separation of concerns with HTML template generation moved to a separate file.

## Project Structure

- `main.go`: The main application file.
- `templates/index.html`: The HTML template for the web page.

## Technologies Used

- **Go**: Programming language for the backend server.
- **Plotly.js**: JavaScript library for interactive charts.
- **HTML & CSS**: For web page structure and styling.
- **JavaScript**: For client-side interactivity.

### Requirements

- Go installed (version 1.16 or higher)
