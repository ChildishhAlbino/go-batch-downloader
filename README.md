# A Go-gram for downloading files in batches
This is very WIP and my first time using Go so I don't know how efficient this is.

## Building
Running `make build` will output both a linux and windows versions to `dist/`

## Running
Running the program will look for a `files.txt` file in the current directory as the program. It will separate each line in the file as a URL and attempt to download them in parallel. 
