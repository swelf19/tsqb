package devapp2

//Store -
//tsqb:gen
//tsqb:tablename=stores
type Store struct {
	ID        int    `tsqb:"col=id"`
	StoreName string `tsqb:"col=storename"`
}
