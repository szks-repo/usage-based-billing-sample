package pipeline

import (
	"context"
	"sync"
	"time"
)

type (
	ProcessFunc[T, U any] func(T) (U, error)

	FilterFunc[T any] func(T) bool

	ForEachFunc[T any] func(T)

	StageFunc[T, U any] func(context.Context, <-chan T) <-chan U
)

// Map: 型安全なデータ変換
func Map[T, U any](fn ProcessFunc[T, U]) StageFunc[T, U] {
	return func(ctx context.Context, input <-chan T) <-chan U {
		output := make(chan U)
		go func() {
			defer close(output)
			for data := range input {
				if result, err := fn(data); err == nil {
					select {
					case output <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
		return output
	}
}

// ParallelMap: 型安全な並列データ変換
func ParallelMap[T, U any](fn ProcessFunc[T, U], workers int) StageFunc[T, U] {
	return func(ctx context.Context, input <-chan T) <-chan U {
		output := make(chan U)
		var wg sync.WaitGroup

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for data := range input {
					if result, err := fn(data); err == nil {
						select {
						case output <- result:
						case <-ctx.Done():
							return
						}
					}
				}
			}(i)
		}

		go func() {
			wg.Wait()
			close(output)
		}()

		return output
	}
}

// Filter: 型安全なフィルタリング
func Filter[T any](predicate FilterFunc[T]) StageFunc[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for data := range input {
				if predicate(data) {
					select {
					case output <- data:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
		return output
	}
}

// ForEach: 副作用実行（型を変更しない）
func ForEach[T any](fn ForEachFunc[T]) StageFunc[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for data := range input {
				fn(data)
				select {
				case output <- data:
				case <-ctx.Done():
					return
				}
			}
		}()
		return output
	}
}

// Delay: 遅延追加
func Delay[T any](d time.Duration) StageFunc[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for data := range input {
				time.Sleep(d)
				select {
				case output <- data:
				case <-ctx.Done():
					return
				}
			}
		}()
		return output
	}
}

// Buffer: バッファリング
func Buffer[T any](size int) StageFunc[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T, size)
		go func() {
			defer close(output)
			for data := range input {
				select {
				case output <- data:
				case <-ctx.Done():
					return
				}
			}
		}()
		return output
	}
}

// Batch: データをスライスにまとめる
func Batch[T any](size int) StageFunc[T, []T] {
	return func(ctx context.Context, input <-chan T) <-chan []T {
		output := make(chan []T)
		go func() {
			defer close(output)
			batch := make([]T, 0, size)

			for data := range input {
				batch = append(batch, data)
				if len(batch) == size {
					select {
					case output <- batch:
						batch = make([]T, 0, size)
					case <-ctx.Done():
						return
					}
				}
			}

			if len(batch) > 0 {
				select {
				case output <- batch:
				case <-ctx.Done():
				}
			}
		}()
		return output
	}
}

// FlatMap: スライスを個別要素に展開
func FlatMap[T any]() StageFunc[[]T, T] {
	return func(ctx context.Context, input <-chan []T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for batch := range input {
				for _, item := range batch {
					select {
					case output <- item:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
		return output
	}
}

// Reduce: アキュムレータパターン
func Reduce[T, U any](initial U, reducer func(U, T) U) StageFunc[T, U] {
	return func(ctx context.Context, input <-chan T) <-chan U {
		output := make(chan U, 1)
		go func() {
			defer close(output)
			acc := initial
			for data := range input {
				acc = reducer(acc, data)
			}
			output <- acc
		}()
		return output
	}
}

// Take: 最初のN個の要素のみ取得
func Take[T any](n int) StageFunc[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			count := 0
			for data := range input {
				if count >= n {
					break
				}
				select {
				case output <- data:
					count++
				case <-ctx.Done():
					return
				}
			}
		}()
		return output
	}
}

// Skip: 最初のN個の要素をスキップ
func Skip[T any](n int) StageFunc[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			count := 0
			for data := range input {
				if count < n {
					count++
					continue
				}
				select {
				case output <- data:
				case <-ctx.Done():
					return
				}
			}
		}()
		return output
	}
}

// パイプライン型定義と実行関数

// Pipeline1: 1つのステージ
func Pipeline1[T, U any](
	ctx context.Context,
	input <-chan T,
	stage1 StageFunc[T, U],
) <-chan U {
	return stage1(ctx, input)
}

// Pipeline2: 2つのステージ
func Pipeline2[T, U, V any](
	ctx context.Context,
	input <-chan T,
	stage1 StageFunc[T, U],
	stage2 StageFunc[U, V],
) <-chan V {
	return stage2(ctx, stage1(ctx, input))
}

// Pipeline3: 3つのステージ
func Pipeline3[T, U, V, W any](
	ctx context.Context,
	input <-chan T,
	stage1 StageFunc[T, U],
	stage2 StageFunc[U, V],
	stage3 StageFunc[V, W],
) <-chan W {
	return stage3(ctx, stage2(ctx, stage1(ctx, input)))
}

// Pipeline4: 4つのステージ
func Pipeline4[T, U, V, W, X any](
	ctx context.Context,
	input <-chan T,
	stage1 StageFunc[T, U],
	stage2 StageFunc[U, V],
	stage3 StageFunc[V, W],
	stage4 StageFunc[W, X],
) <-chan X {
	return stage4(ctx, stage3(ctx, stage2(ctx, stage1(ctx, input))))
}

// Pipeline5: 5つのステージ
func Pipeline5[T, U, V, W, X, Y any](
	ctx context.Context,
	input <-chan T,
	stage1 StageFunc[T, U],
	stage2 StageFunc[U, V],
	stage3 StageFunc[V, W],
	stage4 StageFunc[W, X],
	stage5 StageFunc[X, Y],
) <-chan Y {
	return stage5(ctx, stage4(ctx, stage3(ctx, stage2(ctx, stage1(ctx, input)))))
}

// データソースと結果収集

// From: スライスからチャンネルを作成
func From[T any](data []T) <-chan T {
	output := make(chan T)
	go func() {
		defer close(output)
		for _, item := range data {
			output <- item
		}
	}()
	return output
}

// Collect: チャンネルからスライスに収集
func Collect[T any](ctx context.Context, input <-chan T) []T {
	var results []T
	for data := range input {
		results = append(results, data)
	}
	return results
}

// 型安全な合成関数
type Compose2[T, U, V any] struct {
	stage1 StageFunc[T, U]
	stage2 StageFunc[U, V]
}

func (c Compose2[T, U, V]) Apply(ctx context.Context, input <-chan T) <-chan V {
	return c.stage2(ctx, c.stage1(ctx, input))
}

func NewCompose2[T, U, V any](stage1 StageFunc[T, U], stage2 StageFunc[U, V]) Compose2[T, U, V] {
	return Compose2[T, U, V]{stage1, stage2}
}

// 高度な型安全操作

// Conditional: 型安全な条件分岐
func Conditional[T any](
	predicate FilterFunc[T],
	truePipeline, falsePipeline StageFunc[T, T],
) StageFunc[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for data := range input {
				var result <-chan T
				singleChan := make(chan T, 1)
				singleChan <- data
				close(singleChan)

				if predicate(data) {
					result = truePipeline(ctx, singleChan)
				} else {
					result = falsePipeline(ctx, singleChan)
				}

				for processedData := range result {
					select {
					case output <- processedData:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
		return output
	}
}
