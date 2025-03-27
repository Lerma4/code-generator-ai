This README provides a comprehensive overview of your application, including how to use the TUI functions (in English), the template structure, and an example of the config.json file. It should help users understand how to set up and use your code generator tool.

# Code Generator AI

A terminal-based user interface (TUI) for generating code using AI models.

## Overview

Code Generator AI is a tool that helps developers generate code snippets based on predefined templates. It uses the Gemini API to process prompts and generate code according to your needs.

## Features

- Interactive terminal UI built with Bubble Tea
- Template selection from local directories
- Integration with Gemini AI for code generation
- Configurable settings via JSON

## Installation

1. Clone the repository
2. Ensure Go is installed on your system
3. Run `go build` to compile the application
4. Create a `templates` directory with your code templates
5. Configure your `config.json` file with API keys

## Usage

1. Run the application
2. Select a template using arrow keys
3. Press Enter to generate code based on the selected template
4. Use ESC or Backspace to return to the template selection
5. Press q or Ctrl+C to exit the application

## Template Structure

Each template should be a directory inside the `templates` folder and must contain a `prompt.txt` file with instructions for the AI model.

## Configuration

The application uses a `config.json` file for configuration. Here's an example:

```json
{
  "database": {
    "driver": "mysql",
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "your-password",
    "dbname": "code_generator",
    "max_open_conns": 10,
    "max_idle_conns": 5,
    "conn_max_lifetime": 3600
  },
  "gemini": {
    "model_name": "your-favorite-model",
    "api_key": "your-gemini-api-key"
  }
}
```

## Logs

The application logs are stored in the `logs` directory.