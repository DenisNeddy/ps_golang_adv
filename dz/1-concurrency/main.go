package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	// Инициализируем генератор случайных чисел
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Создаем канал для передачи чисел от первой ко второй горутине

	numbers := make(chan int, 5) // буферизованный канал для небольшой асинхронности
	// Создаем канал для передачи результатов в main
	results := make(chan int, 10)
	//Создаем WaitGroup для ожидания завершения горутин
	var wg sync.WaitGroup

	// Запускаем первую горутину (генератор чисел)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(numbers) // Закрываем канал после отправки всех чисел

		// Создаем slice из 10 случайный чисел
		randomNumbers := make([]int, 10)

		for i := 0; i < 10; i++ {
			randomNumbers[i] = rand.Intn(101) // Числа от 0 до 100
		}

		fmt.Printf("Первая горутина сгенирировала числа: %v\n", randomNumbers)

		// Передаем числа по одному во втрую горутину

		for _, num := range randomNumbers {
			numbers <- num
			fmt.Printf("Отправлено число: %d\n", num)
			time.Sleep(100 * time.Millisecond) // Небольшая задержка для наглядности
		}
	}()

	// Запускаем втроую горутину (вычисление квадратов)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(results) // Закрываем канал результатов после обработки всех чисел

		// Получаем числа из канала и вычисляем квадраты
		for num := range numbers {
			square := num * num
			fmt.Printf("Вторая горутина: %d² = %d\n", num, square)
			results <- square // Отправляем результат в main

		}
	}()

	// Горутина для ожидания завершения рабочих горутин

	go func() {
		wg.Wait()
	}()

	// main собираем все результаты
	fmt.Println("Main: ожидаю результаты...")

	collectedResults := make([]int, 0, 10)
	for result := range results {
		collectedResults = append(collectedResults, result)

		fmt.Printf("Main: получен результат %d\n", result)
	}

	// Выводим все результаты
	fmt.Println("\nВсе результаты (числа в квадрате):")
	for i, result := range collectedResults {
		fmt.Printf("Результат %d: %d\n", i+1, result)
	}

}
