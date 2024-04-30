package Evaluator

/*Quicksort*/
func swap(num1 *int64, num2 *int64) {
	temp := *num1
	*num1 = *num2
	*num2 = temp

}

func partition(arr []int64, low int64, high int64) int64 {
	var pivot int64 = arr[high]
	var i int64 = (low - 1)

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			swap(&arr[i], &arr[j])
		}
	}
	swap(&arr[i+1], &arr[high])
	return (i + 1)
}

func quickSort(arr []int64, low int64, high int64) {

	if low < high {
		var pi int64 = partition(arr, low, high)

		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}

}

/*Quicksort*/
