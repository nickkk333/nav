package service

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nickkk333/nav/database"
	"github.com/nickkk333/nav/logger"
	"github.com/nickkk333/nav/types"
	"github.com/nickkk333/nav/utils"
)

// 首页显示配置的默认值（与前端 utils/theme.ts 保持一致）
const (
	DefaultColorTheme       = "auto"
	DefaultCardShowDuration = 5
	MinCardShowDuration     = 1
	MaxCardShowDuration     = 600
)

// bingWallpaperSubDir 必应每日壁纸固定使用的子目录：
// 壁纸不再单独配置存储目录，统一保存在便签附件存储目录下的「每日壁纸」里
const bingWallpaperSubDir = "每日壁纸"

// 默认存储目录跟随用户文件目录（NAV_FILES_DIR，默认与数据目录一致）：
// 便签附件默认 <文件目录>/memo-files，必应壁纸固定为 <便签附件目录>/每日壁纸
var (
	// DefaultMemoFileDir 便签附件的默认存储目录
	DefaultMemoFileDir = filepath.Join(utils.FilesDir(), "memo-files")
	// DefaultBingWallpaperDir 必应每日壁纸的默认存储目录
	DefaultBingWallpaperDir = filepath.Join(DefaultMemoFileDir, bingWallpaperSubDir)
)

// defaultSetting 返回读取不到数据库时的默认设置
func defaultSetting() types.Setting {
	return types.Setting{
		Id:               1,
		Favicon:          "favicon.ico",
		Title:            "Nav",
		GovRecord:        "",
		Logo192:          "logo192.png",
		Logo512:          "logo512.png",
		HideAdmin:        false,
		HideGithub:       false,
		HideMemo:         false,
		HideCatelogNav:   false,
		JumpTargetBlank:  true,
		BackgroundImage:  "",
		DefaultLogo:      "",
		ColorTheme:       DefaultColorTheme,
		CardShowDuration: DefaultCardShowDuration,
	}
}

// GetSetting 获取网站设置
func GetSetting() types.Setting {
	setting := defaultSetting()
	// 单元测试等尚未初始化数据库的场景直接返回默认值，避免空指针
	if database.DB == nil {
		return setting
	}
	row := database.DB.QueryRow(`
		SELECT id, favicon, title, govRecord, logo192, logo512, hideAdmin, hideGithub, hideMemo, hideCatelogNav, jumpTargetBlank, backgroundImage, defaultLogo, defaultSearchEngine, colorTheme, cardShowDuration, memoFileDir
		FROM nav_setting
		ORDER BY id ASC
		LIMIT 1;
		`)
	var (
		hideAdmin           sql.NullBool
		hideGithub          sql.NullBool
		hideMemo            sql.NullBool
		hideCatelogNav      sql.NullBool
		jumpTargetBlank     sql.NullBool
		backgroundImage     sql.NullString
		defaultLogo         sql.NullString
		defaultSearchEngine sql.NullInt64
		colorTheme          sql.NullString
		cardShowDuration    sql.NullInt64
		memoFileDir         sql.NullString
	)
	err := row.Scan(&setting.Id, &setting.Favicon, &setting.Title, &setting.GovRecord, &setting.Logo192, &setting.Logo512,
		&hideAdmin, &hideGithub, &hideMemo, &hideCatelogNav, &jumpTargetBlank, &backgroundImage, &defaultLogo, &defaultSearchEngine,
		&colorTheme, &cardShowDuration, &memoFileDir)
	if err != nil {
		logger.LogError("获取配置失败: %s", err)
		return setting
	}
	setting.HideAdmin = hideAdmin.Bool
	setting.HideGithub = hideGithub.Bool
	// 未设置时首页显示便签入口
	setting.HideMemo = hideMemo.Bool
	// 未设置时首页显示顶部分类导航
	setting.HideCatelogNav = hideCatelogNav.Bool
	// 未设置时默认新标签页打开
	setting.JumpTargetBlank = true
	if jumpTargetBlank.Valid {
		setting.JumpTargetBlank = jumpTargetBlank.Bool
	}
	// 未设置时前台使用必应每日壁纸作为背景图
	setting.BackgroundImage = backgroundImage.String
	// 未设置默认图标时前台用名称首字符占位（LogoFallback）
	setting.DefaultLogo = defaultLogo.String
	// 未设置默认搜索引擎时回落到「自动」（第一个启用的引擎）
	setting.DefaultSearchEngine = 0
	if defaultSearchEngine.Valid {
		setting.DefaultSearchEngine = int(defaultSearchEngine.Int64)
	}
	// 未设置 / 非法配色时回落到自动主题
	setting.ColorTheme = DefaultColorTheme
	if colorTheme.Valid && (colorTheme.String == "auto" || colorTheme.String == "light" || colorTheme.String == "dark") {
		setting.ColorTheme = colorTheme.String
	}
	// 未设置 / 越界的持续显示时间回落到默认值
	setting.CardShowDuration = DefaultCardShowDuration
	if cardShowDuration.Valid && cardShowDuration.Int64 >= MinCardShowDuration && cardShowDuration.Int64 <= MaxCardShowDuration {
		setting.CardShowDuration = int(cardShowDuration.Int64)
	}
	// 未设置时使用默认存储目录
	setting.MemoFileDir = memoFileDir.String
	return setting
}

// UpdateSetting 更新网站设置
func UpdateSetting(data types.Setting) error {
	// 存储目录先做校验与规范化：空字符串表示使用默认目录，非空时自动创建
	memoFileDir, err := normalizeStorageDir(data.MemoFileDir)
	if err != nil {
		return fmt.Errorf("便签附件存储目录不可用：%w", err)
	}
	data.MemoFileDir = memoFileDir
	// 壁纸固定保存在便签附件目录下的「每日壁纸」子目录，配置了便签附件目录时一并创建
	if memoFileDir != "" {
		if err := os.MkdirAll(filepath.Join(memoFileDir, bingWallpaperSubDir), 0o755); err != nil {
			return fmt.Errorf("创建每日壁纸目录失败：%w", err)
		}
	}
	_, err = database.DB.Exec(`
		UPDATE nav_setting
		SET favicon = ?, title = ?, govRecord = ?, logo192 = ?, logo512 = ?, hideAdmin = ?, hideGithub = ?, hideMemo = ?, hideCatelogNav = ?, jumpTargetBlank = ?, backgroundImage = ?, defaultLogo = ?, defaultSearchEngine = ?, colorTheme = ?, cardShowDuration = ?, memoFileDir = ?
		WHERE id = (SELECT id FROM nav_setting ORDER BY id ASC LIMIT 1);
		`, data.Favicon, data.Title, data.GovRecord, data.Logo192, data.Logo512,
		data.HideAdmin, data.HideGithub, data.HideMemo, data.HideCatelogNav, data.JumpTargetBlank, data.BackgroundImage, data.DefaultLogo,
		data.DefaultSearchEngine, data.ColorTheme, data.CardShowDuration, data.MemoFileDir)
	return err
}

// normalizeStorageDir 校验并规范化存储目录：
// 空字符串表示使用默认目录；非空时必须是目录（不存在则创建）
func normalizeStorageDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", nil
	}
	dir = filepath.Clean(dir)
	if info, err := os.Stat(dir); err == nil && !info.IsDir() {
		return "", fmt.Errorf("%s 已存在且不是目录", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建目录失败：%w", err)
	}
	return dir, nil
}

// resolveStorageDir 把配置值解析成实际目录，留空时回落到默认目录
func resolveStorageDir(configured, fallback string) string {
	dir := strings.TrimSpace(configured)
	if dir == "" {
		return fallback
	}
	return filepath.Clean(dir)
}

// MemoFileDir 返回当前生效的便签附件存储目录
func MemoFileDir() string {
	if database.DB == nil {
		return DefaultMemoFileDir
	}
	return resolveStorageDir(GetSetting().MemoFileDir, DefaultMemoFileDir)
}

// BingWallpaperDir 返回当前生效的必应每日壁纸存储目录：
// 固定在便签附件存储目录下的「每日壁纸」子目录，不再单独配置
func BingWallpaperDir() string {
	if database.DB == nil {
		return DefaultBingWallpaperDir
	}
	return filepath.Join(MemoFileDir(), bingWallpaperSubDir)
}
