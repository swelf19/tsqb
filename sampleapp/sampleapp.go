/* В devapp мы разрабатываем наш квери билдер и тесты к нему/

Тесты и структуры копируем в sampleapp из devapp,
кверибилдер для sampleapp генерируем из структур
*/

package sampleapp

//User -
//tsqb:gen
//tsqb:tablename=users
type User struct {
	ID       int    `tsqb:"col=id"`
	UserName string `tsqb:"col=username"`
	LastLog  string `tsqb:"col=last_log"`
}

//Store -
//tsqb:gen
//tsqb:tablename=stores
type Store struct {
	ID        int    `tsqb:"col=id"`
	StoreName string `tsqb:"col=storename"`
}
