# Real-Time Currency Rates

A Go application that displays real-time currency exchange rates for **USD to EUR** and **BTC to USD**. The rates are updated every 10 seconds and presented on interactive charts.

## Project Description

This web application fetches current exchange rates from public APIs and displays them on interactive charts using Plotly.js. The charts are arranged horizontally:

- **Left**: USD to EUR exchange rate chart.
- **Right**: BTC to USD exchange rate chart.

### Key Features

- **Real-Time Updates**: Fetches new data every 10 seconds.
- **Interactive Charts**: Users can zoom, pan, and hover over data points to see detailed information.
- **Fixed Y-Axis Range**: The USD to EUR chart's y-axis is fixed between 0.7 and 1.2 for better visualization.
- **Responsive Design**: Charts are arranged horizontally for easy comparison.
- **Error Handling**: Robust error handling and logging.
- **Clean Architecture**: Separation of concerns with the HTML template in a separate file.
- **English Language**: All code, comments, labels, and titles are in English for consistency.

## Project Structure

- `main.go`: The main application file containing the Go code.
- `templates/index.html`: The HTML template for the web page.

## Technologies Used

- **Go**: Programming language for the backend server.
- **Plotly.js**: JavaScript library for interactive charts.
- **HTML & CSS**: For web page structure and styling.
- **JavaScript**: For client-side interactivity.
