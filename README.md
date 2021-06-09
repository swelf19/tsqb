# TSQB - TypeSafeQueryBuilder

## Цель
Proof of concept реализация подмножества sql запросов на языке golang, работающий на основе кодогенерации по заранее заданной схеме.
В данный момент реализованны простые операции select,update,delete,insert. Без join'ов, без аггрегационных функций.

## Использование
### Описываем нашу схему
Для начала необходимо описать нашу схему - обычную структуру golang
```go
//schema.go

//tsqb:gen
//tsqb:tablename=users
type User struct {
	ID       int    `tsqb:"col=id"`
	UserName string `tsqb:"col=username"`
	LastLog  string `tsqb:"col=last_log"`
}
```
//tsqb:gen - директива для принятия решения парсером, надо ли генерировать код для этой структуры
//tsqb:tablename=users - имя таблицы использующееся в sql запросах этой схемы

struct tags `tsqb:"col=id"` используются для связи полей структуры со столбцами sql таблицы

### Генерируем код
Далее запускаем кодогенератор
go run github.com/swelf19/tsqb/tsqb ./schema.go
можно передать директорию, утилита рекурсивной обоядет все файлы в поисках директивы
//tsqb:gen
пример сгенерированного кода - https://github.com/swelf19/tsqb/blob/master/devapp2/devapp_gen.go
После чего, код можно использовать.

### Написание запросов

#### SELECT
```go
//Создаем query builder для схемы User
bu := Select().User()
//Больше можно ничего не делать, вызываем метод Build() который подготовит наш builder для дальнейшего использования, а метод SQL() вернет нам строку с запросом
fmt.Println(bu.Build().SQL())
//select users.id, users.username, users.last_log from users

//Добавим пару условий limit,offset
fmt.Println(bu.Limit(20).Offset(10).Build().SQL())
//select users.id, users.username, users.last_log from users offset 10 limit 20

//Добавим пару условий where. Условия создаются при помощи ранее созданного builder'а, В качестве аргумента, функция вызываемая для создания where условия, принимает тип описанный в "схеме" данных
Query := bu.Where(
		bu.UserSchema.Fields.ID.Eq(1),
		bu.UserSchema.Fields.ID.Eq(2),
	).Build()
fmt.Println(Query.SQL())
//select users.id, users.username, users.last_log from users where (users.id = $1 and users.id = $2)
//генерируется postgres специфичные запросы с параметрами для prepared statements
//Сами значения параметров хранятся во внутренней структуре типа []interface{}

Query = b.Where(
		b.UserSchema.Fields.ID.Eq(1),
		qfuncs.ComposeOr(
			b.UserSchema.Fields.UserName.Eq("swelf"),
			b.UserSchema.Fields.UserName.Eq("admin"),
			qfuncs.ComposeAnd(
				b.UserSchema.Fields.UserName.Eq("lalala"),
				b.UserSchema.Fields.UserName.Eq("lalala2"),
			),
			qfuncs.ComposeAnd(
				b.UserSchema.Fields.UserName.Eq("lalala3"),
				b.UserSchema.Fields.UserName.Eq("lalala4"),
			),
		),
	).Build()
//Так же можно строить запросы любой вложенности объеденяя условия при помощи функций qfuncs.ComposeOr и qfuncs.ComposeAnd
//select users.id, users.username, users.last_log from users where (users.id = $1 and (users.username = $2 or users.username = $3 or (users.username = $4 and users.username = $5) or (users.username = $6 and users.username = $7)))

//В данный момент в коде зашито использование драявера pgx для выполнения запросов.
conn, _ := pgx.Connect(context.Background(), dbURL)
items, err := Query.Fetch(context.Background(), conn)
//Метод Fetch вернет нам []User
```

#### INSERT
```go
u := User{
		ID:       0,
		UserName: "swelf",
		LastLog:  "today",
	}
b := Insert().User(u)
//insert into users(username, last_log) values($1, $2) returning id
q := b.Build()
id, err := q.Exec(context.Background(), conn)
//Так же как и с операцией select, мы можем выполнить запрос при помощи драйвера pgx, метод exec вернет нам ID новой сущности
```

#### UPDATE
```go
b := Update().User()
//Поддерживается 2 вида update запросов.
// 1 - передаем в качестве аргументов всю вструктуру
Query := b.SetAllFields(u).Build()
//update users set username = $2, last_log = $3 where users.id = $1

//2 - передаем каждое поле отдально
Query := b.SetUserName("lala").SetLastLog("tomorrow").Where(b.UserSchema.Fields.ID.Eq(1)).Build()
//update users set username = $2, last_log = $3 where users.id = $1
//Нужно обратить внимание, если мы не задаем условие Where в данном запросе, то у нас обновятся поля во всей таблице
Query := b.SetUserName("lala").SetLastLog("tomorrow").Build()
//update users set username = $1, last_log = $2
```

#### DELETE
```go
b := Delete().User()
Query := b.Build()
//delete from users
Query = b.Where(b.UserSchema.Fields.ID.Eq(1)).Build()
//delete from users where users.id = $1
```
