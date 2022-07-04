/* Инициализация логера приложения
 */

package logger

import (
	"log"
	"os"
)

/* Создание логгеров уровня info и error
 */
func GetLogger() (*log.Logger, *log.Logger) {
	return log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime), log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
}
