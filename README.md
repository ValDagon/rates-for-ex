# Rate For Ex Version 0.8

A Go application that displays real-time currency exchange rates for **USD to EUR** and **BTC to USD**. The rates are updated every 10 seconds and presented on interactive charts.

## Overview

Go-based web application that fetches and displays real-time currency exchange rates for USD to EUR and BTC to USD. The application features dynamic charts powered by Plotly, real-time status indicators, and customizable UI elements to enhance readability and user experience.

## Features

- **Real-Time Data Fetching:**

  - **USD to EUR:** Scrapes exchange rates from [x-rates.com](https://www.x-rates.com/).
  - **BTC to USD:** Fetches exchange rates using the [CoinDesk API](https://api.coindesk.com/v1/bpi/currentprice/USD.json).

- **Dynamic Charts:**

  - Utilizes Plotly for interactive and responsive chart visualizations.
  - Customizable font sizes, colors, and margins to prevent label overlap.

- **Status Indicators:**

  - Displays current application status ("All is working" or "Error fetching data") to inform users of operational state.

- **Version Information:**
  - Embeds version details into the binary for easy tracking.

## Project Structure

- `main.go`: The main application file containing the Go code.
- `Makefile`: Facilitates easy building and cleaning of the binary.
- `templates/index.html`: The HTML template for the web page.

## Technologies Used

- **Go**: Programming language for the backend server.
- **Plotly.js**: JavaScript library for interactive charts.
- **HTML & CSS**: For web page structure and styling.
- **JavaScript**: For client-side interactivity.

## **Building the Binary**

1. Using Go Build Command:

   - Navigate to the project directory:
     cd /path/to/your/project
   - Build the binary with embedded version information:
     go build -ldflags "-X main.version=0.8" -o "R-F-E v0.8"

2. Using Makefile:
   - Ensure the Makefile contains the appropriate build instructions.
   - Run the build command:
     make build
   - To clean the binary:
     make clean

## **Running the Application**

1. Ensure execute permissions:
   chmod +x "R-F-E v0.8"

2. Run the binary:
   ./R-F-E\ v0.8
   or
   ./"R-F-E v0.8"

3. Access the web interface:
   Open your web browser and navigate to http://localhost:8080 to view the real-time exchange rates and charts.

## **Application Flow**

1. **Data Fetching**:

   - The application starts a goroutine that fetches exchange rates every 10 seconds.
   - USD to EUR: Scrapes data from x-rates.com.
   - BTC to USD: Fetches data from the CoinDesk API.

2. **Data Storage**:

   - Stores fetched data points in a slice, retaining only the last hour of data to optimize performance.

3. **Web Server**:

   - Serves the main page (index.html) displaying the charts and current rates.
   - Provides a /data endpoint that returns the latest data in JSON format for real-time updates.

4. **Chart Rendering**:

   - Utilizes Plotly to render dynamic and interactive charts.
   - Configured with increased font sizes, bold and black text, and optimized margins to prevent label overlap.

5. **Status Indicators**:

   - Displays the current operational status of the application, indicating whether data fetching is successful or if errors have occurred.

6. **Graceful Shutdown**:
   - Listens for termination signals (e.g., Ctrl+C) to gracefully shut down the server, ensuring all resources are properly released.

## Customizing the UI

- **Font Customizations:**

  - Increase the `font-size` properties in CSS to make text more prominent.
  - Set `font-weight: bold` and `color: #000` to ensure text is both bold and black.

- **Margin Adjustments:**

  - Increase the `margin` properties in the Plotly layout configurations to provide more space around the charts, preventing labels from clashing with chart boundaries.

- **Status Indicators:**
  - The status indicator at the bottom of the page changes color and text based on the application's state, providing immediate feedback to users.

## Troubleshooting

- **Version Not Displayed:**

  - Ensure the `-ldflags` flag correctly sets the `version` variable.
  - Verify that `main.go` references the `version` variable appropriately.

- **Data Fetching Errors:**

  - Check network connectivity to x-rates.com and the CoinDesk API.
  - Monitor application logs for error messages related to data fetching.

- **Chart Rendering Issues:**

  - Ensure that Plotly.js is correctly loaded.
  - Verify that the data passed to Plotly is correctly formatted.

- **Server Not Starting:**
  - Confirm that port `8080` is not in use by another application.
  - Check application logs for any startup errors.

## Enhancements and Future Work

- **Responsive Design:** Implement media queries to ensure the application looks good on various screen sizes.
- **Extended Data Retention:** Adjust the data retention policy based on user needs, potentially allowing for longer historical data.
- **Additional Exchange Rates:** Expand the application to include more currency pairs or cryptocurrencies.
- **Deployment:** Consider containerizing the application using Docker for easier deployment and scalability.
