package main

import (
	"errors"
	"fmt"
	"os"
)

type TaskStatus string

type Task struct {
	name       string
	maxRetries int
	status     TaskStatus
}

const (
	Pending   TaskStatus = "pending"
	Running   TaskStatus = "running"
	Retrying  TaskStatus = "retrying"
	Succeeded TaskStatus = "succeeded"
	Failed    TaskStatus = "failed"
)

func toolCall() error {
	return errors.New("temporary tool failure")
}

func printTaskStatus(current int, task Task, err error) {
	switch task.status {
	case Retrying:
		fmt.Printf(
			"任务进行第 %d 次重试，失败原因 %s \n",
			current,
			err,
		)

	case Failed:
		fmt.Printf("任务 %s 失败，失败原因 %s \n", task.name, err)
	case Running:
		fmt.Println("任务执行中")
	case Succeeded:
		fmt.Println("任务成功")
	}

}

func runTask(task *Task) error {

	for current := 1; current <= task.maxRetries; current++ {
		task.status = Running
		printTaskStatus(current, *task, nil)
		err := toolCall()

		if err == nil {
			task.status = Succeeded
			printTaskStatus(current, *task, err)
			return nil
		}

		if current < task.maxRetries {
			task.status = Retrying
			printTaskStatus(current, *task, err)
			continue
		}

		task.status = Failed

		printTaskStatus(current, *task, err)

		return err

	}

	return nil
}

func main() {
	name := os.Getenv("TASK_NAME")

	task := Task{
		name:       name,
		maxRetries: 3,
		status:     Pending,
	}

	if task.name == "" {
		fmt.Println("任务名称不能为空")
		return
	}

	if task.maxRetries <= 0 {
		fmt.Println("重试次数非法")
		return
	}

	taskErr := runTask(&task)

	if taskErr != nil {
		return
	}

	fmt.Println("任务执行完成")
}
