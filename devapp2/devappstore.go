package devapp2

//Store -

type Store struct {
	ID        int    `tsqb:"col=id"`
	StoreName string `tsqb:"col=storename"`
}
