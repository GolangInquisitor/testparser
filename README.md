# testparser
Тестовое задание

#### Пример использования

```
package main

  import (
  
  "fmt"
  "github.com/GolangInquisitor/testparser/ymcparser"
 
  )
  
  const YouGetGeoApiKey = "Здесь надо вбить APIkey от Google Geocoder"
  
  func main() {
  
	fmt.Println("Start Parsing!")
  
	var ymcprser ymcparser.Parser

	ymcprser.Run(YouGetGeoApiKey)
  
  }
  ```
