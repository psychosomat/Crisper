package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"Crisper/internal/config"
	"Crisper/internal/hardware"
	"Crisper/internal/models"
	"Crisper/internal/queue"
	"Crisper/internal/transcribe"
	"Crisper/internal/whisper"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	config     config.Settings
	queue      *queue.Queue
	mu         sync.Mutex
	cancelFn   context.CancelFunc
	done       chan struct{}
}

func NewApp() *App {
	cfg, _ := config.Load()
	q := queue.New()
	return &App{
		config: cfg,
		queue:  q,
		done:   make(chan struct{}),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.queue.SetUpdateCallback(func() {
		runtime.EventsEmit(a.ctx, "queue-update")
	})
}

func (a *App) shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancelFn != nil {
		a.cancelFn()
		a.cancelFn = nil
	}
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	a.shutdown(ctx)
	return false
}

func (a *App) GetSettings() config.Settings {
	return a.config
}

func (a *App) SaveSettings(s config.Settings) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config = s
	return config.Save(s)
}

func (a *App) GetHardwareSpecs() hardware.Specs {
	return hardware.Detect()
}

func (a *App) GetAvailableModels() []models.ModelInfo {
	return models.AllModels()
}

func (a *App) GetRecommendedModel() models.ModelInfo {
	return models.RecommendModel()
}

func (a *App) IsModelDownloaded(name string) bool {
	storageDir := config.ModelsDir()
	return models.IsDownloaded(storageDir, name)
}

func (a *App) DownloadModel(name string) error {
	var target models.ModelInfo
	found := false
	for _, m := range models.AllModels() {
		if m.Name == name {
			target = m
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unknown model: %s", name)
	}

	storageDir := config.ModelsDir()
	if models.IsDownloaded(storageDir, name) {
		return nil
	}

	err := models.DownloadWithProgress(storageDir, target, func(downloaded, total int64) {
		progress := float64(0)
		if total > 0 {
			progress = float64(downloaded) / float64(total)
		}
		runtime.EventsEmit(a.ctx, "download-progress", map[string]interface{}{
			"model":      name,
			"downloaded": downloaded,
			"total":      total,
			"progress":   progress,
		})
	})

	if err != nil {
		return err
	}

	a.mu.Lock()
	a.config.ModelName = name
	a.mu.Unlock()
	return config.Save(a.config)
}

func (a *App) IsWhisperInstalled() bool {
	_, err := whisper.FindWhisperCLI()
	return err == nil
}

func (a *App) GetWhisperPath() string {
	p, _ := whisper.FindWhisperCLI()
	return p
}

func (a *App) DownloadWhisperCLI() error {
	binDir := config.WhisperBinDir()
	path, err := whisper.DownloadWhisperCLI(binDir, func(downloaded, total int64) {
		progress := float64(0)
		if total > 0 {
			progress = float64(downloaded) / float64(total)
		}
		runtime.EventsEmit(a.ctx, "whisper-download-progress", map[string]interface{}{
			"downloaded": downloaded,
			"total":      total,
			"progress":   progress,
		})
	})
	if err != nil {
		return err
	}

	whisper.InvalidateCache()

	a.mu.Lock()
	a.config.WhisperPath = path
	a.mu.Unlock()
	return config.Save(a.config)
}

func (a *App) AddFiles(paths []string) []queue.TaskInfo {
	absPaths := make([]string, len(paths))
	for i, p := range paths {
		abs, err := filepath.Abs(p)
		if err == nil {
			absPaths[i] = abs
		} else {
			absPaths[i] = p
		}
	}
	tasks := a.queue.AddFiles(absPaths)
	out := make([]queue.TaskInfo, len(tasks))
	for i, t := range tasks {
		out[i] = *t
	}
	return out
}

func (a *App) RemoveFile(id string) {
	a.queue.Remove(id)
}

func (a *App) ClearQueue() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.queue = queue.New()
	a.queue.SetUpdateCallback(func() {
		runtime.EventsEmit(a.ctx, "queue-update")
	})
	runtime.EventsEmit(a.ctx, "queue-update")
}

func (a *App) GetQueue() []queue.TaskInfo {
	return a.queue.Tasks()
}

func (a *App) IsProcessing() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cancelFn != nil
}

func (a *App) StartQueue(modelName string, outputDir string) error {
	a.mu.Lock()
	if a.cancelFn != nil {
		a.mu.Unlock()
		return fmt.Errorf("already processing")
	}

	if modelName == "" {
		a.mu.Unlock()
		return fmt.Errorf("no model selected")
	}

	storageDir := config.ModelsDir()
	modelFile := filepath.Join(storageDir, modelFileName(modelName))
	if _, err := os.Stat(modelFile); os.IsNotExist(err) {
		a.mu.Unlock()
		return fmt.Errorf("model not downloaded: %s", modelName)
	}

	cli, err := whisper.FindWhisperCLI()
	if err != nil {
		a.mu.Unlock()
		binDir := config.WhisperBinDir()
		_, dlErr := whisper.DownloadWhisperCLI(binDir, func(downloaded, total int64) {
			progress := float64(0)
			if total > 0 {
				progress = float64(downloaded) / float64(total)
			}
			runtime.EventsEmit(a.ctx, "whisper-download-progress", map[string]interface{}{
				"downloaded": downloaded,
				"total":      total,
				"progress":   progress,
			})
		})
		if dlErr != nil {
			return fmt.Errorf("whisper-cli auto-download failed: %s", dlErr.Error())
		}
		whisper.InvalidateCache()

		a.mu.Lock()
		if a.cancelFn != nil {
			a.mu.Unlock()
			return fmt.Errorf("already processing")
		}
		cli, err = whisper.FindWhisperCLI()
		if err != nil {
			a.mu.Unlock()
			return fmt.Errorf("whisper-cli failed to install: %w", err)
		}
	}
	_ = cli

	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelFn = cancel
	a.mu.Unlock()

	m := whisper.NewModel(modelFile, modelName)

	tmpDir, _ := filepath.Abs(filepath.Join(config.ModelsDir(), "..", "tmp"))

	cfg := transcribe.Config{
		Model:          m,
		Language:       a.config.Language,
		Threads:        a.config.Threads,
		ShowTimestamps: a.config.ShowTimestamps,
		OutputDir:      outputDir,
		TmpDir:         tmpDir,
	}

	go a.runLoop(ctx, cfg)

	return nil
}

func (a *App) runLoop(ctx context.Context, cfg transcribe.Config) {
	defer func() {
		a.mu.Lock()
		a.cancelFn = nil
		a.mu.Unlock()
		runtime.EventsEmit(a.ctx, "queue-update")
	}()

	processed := make(map[string]bool)
	doneCount := 0

	for {
		tasks := a.queue.Tasks()
		active := false

		for _, task := range tasks {
			if task.Status == queue.StatusDone || task.Status == queue.StatusError || task.Status == queue.StatusPaused {
				continue
			}

			if processed[task.ID] {
				active = true
				continue
			}

			select {
			case <-ctx.Done():
				return
			default:
			}

			a.queue.UpdateTask(task.ID, queue.StatusProcessing, 0, "")

			_, err := transcribe.Run(ctx, task.FilePath, cfg, func(phase string, progress float64) {
				a.queue.UpdateTask(task.ID, queue.StatusProcessing, progress, "")
				runtime.EventsEmit(a.ctx, "task-progress", map[string]interface{}{
					"id":       task.ID,
					"phase":    phase,
					"progress": progress,
				})
			})

			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					a.queue.UpdateTask(task.ID, queue.StatusPaused, 0, "")
				} else {
					a.queue.UpdateTask(task.ID, queue.StatusError, 0, err.Error())
				}
				processed[task.ID] = true
				runtime.EventsEmit(a.ctx, "queue-update")
				continue
			}

			doneCount++
			a.queue.UpdateTask(task.ID, queue.StatusDone, 1.0, "")
			runtime.EventsEmit(a.ctx, "queue-update")
			a.queue.Remove(task.ID)
			runtime.EventsEmit(a.ctx, "queue-update")
		}

		if !active {
			if doneCount > 0 {
				runtime.EventsEmit(a.ctx, "batch-complete", map[string]interface{}{
					"processed": doneCount,
					"errors":    len(a.queue.Tasks()),
				})
			}
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (a *App) PauseQueue() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancelFn != nil {
		a.cancelFn()
		a.cancelFn = nil
	}
}

func (a *App) CancelQueue() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancelFn != nil {
		a.cancelFn()
		a.cancelFn = nil
	}
	tasks := a.queue.Tasks()
	for _, t := range tasks {
		if t.Status == queue.StatusProcessing || t.Status == queue.StatusPending {
			a.queue.UpdateTask(t.ID, queue.StatusPaused, 0, "")
		}
	}
}

func (a *App) SelectOutputDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select output directory for transcripts",
	})
	if err != nil {
		return "", err
	}
	if dir != "" {
		a.mu.Lock()
		a.config.OutputDir = dir
		a.mu.Unlock()
		config.Save(a.config)
	}
	return dir, nil
}

func (a *App) SelectFiles() ([]string, error) {
	files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select video files",
		Filters: []runtime.FileFilter{
			{DisplayName: "Video Files (*.mp4,*.mkv,*.avi,*.mov,*.webm,*.flv,*.wmv,*.ts)", Pattern: "*.mp4;*.mkv;*.avi;*.mov;*.webm;*.flv;*.wmv;*.ts"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (a *App) SetWindowFrame(frame string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch frame {
	case "none":
		a.config.WindowFrame = "none"
	case "custom":
		a.config.WindowFrame = "custom"
	default:
		a.config.WindowFrame = "system"
	}
	config.Save(a.config)
}

func (a *App) WindowMinimize() {
	runtime.WindowMinimise(a.ctx)
}

func (a *App) WindowToggleMaximize() {
	runtime.WindowToggleMaximise(a.ctx)
}

func (a *App) WindowClose() {
	runtime.Quit(a.ctx)
}

func modelFileName(name string) string {
	for _, m := range models.AllModels() {
		if m.Name == name {
			return m.Filename
		}
	}
	return name
}

