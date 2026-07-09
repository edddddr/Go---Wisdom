package main


import "fmt"


// func add(x, y int) int {
// 	return x + y

// }


func swap(x, y string) (string, string){
	return y, x
}




func main(){


	// fmt.Println(add(10, 10))

	fmt.Println(swap("hello", "worlds"))

}