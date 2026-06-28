package i18n

import (
	"fmt"
	"strings"
	"time"

	"kiri/internal/model"
	"kiri/internal/version"
)

type Catalog struct {
	Lang Lang
}

type FooterHint struct {
	Key   string
	Label string
}

func (c Catalog) footerTabSwitchHint() FooterHint {
	if c.Lang == EN {
		return FooterHint{Key: "⌨ Tab/⇧Tab", Label: "tabs"}
	}
	return FooterHint{Key: "⌨ Tab/⇧Tab", Label: "вкладки"}
}

func New(lang Lang) Catalog {
	return Catalog{Lang: lang}
}

func (c Catalog) ErrorQuit(err error) string {
	if c.Lang == EN {
		return fmt.Sprintf("Error: %v\nPress q to quit.", err)
	}
	return fmt.Sprintf("Ошибка: %v\nНажмите q для выхода.", err)
}

func (c Catalog) TabAllPlants() string {
	if c.Lang == EN {
		return "All Plants"
	}
	return "Все растения"
}

func (c Catalog) TabCalendar(count int) string {
	if count > 0 {
		if c.Lang == EN {
			return fmt.Sprintf("Calendar (%d)", count)
		}
		return fmt.Sprintf("Календарь (%d)", count)
	}
	if c.Lang == EN {
		return "Calendar"
	}
	return "Календарь"
}

func (c Catalog) TabCareLog() string {
	if c.Lang == EN {
		return "Care Log"
	}
	return "Журнал ухода"
}

func (c Catalog) TabSettings() string {
	if c.Lang == EN {
		return "Settings"
	}
	return "Настройки"
}

func (c Catalog) TabAbout() string {
	if c.Lang == EN {
		return "About"
	}
	return "О программе"
}

func (c Catalog) FooterAboutHints() []FooterHint {
	if c.Lang == EN {
		return []FooterHint{
			{Key: "Tab", Label: "switch tabs"},
			{Key: "1–5", Label: "jump to tab"},
			{Key: "q", Label: "quit"},
		}
	}
	return []FooterHint{
		{Key: "  ⌨ Tab", Label: "переключить вкладку"},
		{Key: "1–5", Label: "перейти к вкладке"},
		{Key: "q", Label: "выход"},
	}
}

func (c Catalog) WeatherHeaderLive(cityName string, tempC float64, summary string) string {
	return fmt.Sprintf("%s: %.1f°C, %s", cityName, tempC, summary)
}

func (c Catalog) WeatherLoading(cityName string) string {
	if c.Lang == EN {
		return cityName + ": loading..."
	}
	return cityName + ": загрузка..."
}

func (c Catalog) WeatherUnavailable(cityName string) string {
	if c.Lang == EN {
		return cityName + ": unavailable"
	}
	return cityName + ": недоступна"
}

func (c Catalog) WeatherHeaderDefaultCity() string {
	if c.Lang == EN {
		return "City"
	}
	return "Город"
}

func (c Catalog) WeatherLoadFailed() string {
	if c.Lang == EN {
		return "Unable to load weather right now."
	}
	return "Не удалось получить погоду."
}

func (c Catalog) WeatherCodeLabel(code int) string {
	if c.Lang == EN {
		switch code {
		case 0:
			return "clear"
		case 1:
			return "mainly clear"
		case 2:
			return "partly cloudy"
		case 3:
			return "overcast"
		case 45, 48:
			return "fog"
		case 51:
			return "light drizzle"
		case 53:
			return "drizzle"
		case 55:
			return "heavy drizzle"
		case 56, 57:
			return "freezing drizzle"
		case 61:
			return "light rain"
		case 63:
			return "rain"
		case 65:
			return "heavy rain"
		case 66, 67:
			return "freezing rain"
		case 71:
			return "light snow"
		case 73:
			return "snow"
		case 75:
			return "heavy snow"
		case 77:
			return "snow grains"
		case 80:
			return "light rain showers"
		case 81:
			return "rain showers"
		case 82:
			return "violent showers"
		case 85, 86:
			return "snow showers"
		case 95:
			return "thunderstorm"
		case 96, 99:
			return "thunderstorm with hail"
		default:
			return "unknown"
		}
	}
	switch code {
	case 0:
		return "ясно"
	case 1:
		return "преимущественно ясно"
	case 2:
		return "переменная облачность"
	case 3:
		return "пасмурно"
	case 45, 48:
		return "туман"
	case 51:
		return "слабая морось"
	case 53:
		return "морось"
	case 55:
		return "сильная морось"
	case 56, 57:
		return "ледяная морось"
	case 61:
		return "слабый дождь"
	case 63:
		return "дождь"
	case 65:
		return "сильный дождь"
	case 66, 67:
		return "ледяной дождь"
	case 71:
		return "слабый снег"
	case 73:
		return "снег"
	case 75:
		return "сильный снег"
	case 77:
		return "снежные зерна"
	case 80:
		return "слабые ливни"
	case 81:
		return "ливни"
	case 82:
		return "сильные ливни"
	case 85, 86:
		return "снежные заряды"
	case 95:
		return "гроза"
	case 96, 99:
		return "гроза с градом"
	default:
		return "неизвестно"
	}
}

func (c Catalog) SplashPlantStatusBoard() string {
	if c.Lang == EN {
		return "Plant Status Board"
	}
	return "Панель статуса растений"
}

const aboutGitHubURL = "github.com/freislot/kiri"

func aboutAlignLabelValue(label, value string, labelWidth int) string {
	pad := labelWidth - len([]rune(label))
	if pad < 0 {
		pad = 0
	}
	return label + strings.Repeat(" ", pad) + " " + value
}

func (c Catalog) aboutContactLines() []string {
	labels := []string{"Автор:", "Контакты:", "GitHub:"}
	values := []string{"Павел Антонов", "freislot@gmail.com", aboutGitHubURL}
	if c.Lang == EN {
		labels = []string{"Author:", "Contact:", "GitHub:"}
		values = []string{"Pavel Antonov", "freislot@gmail.com", aboutGitHubURL}
	}
	maxW := 0
	for _, label := range labels {
		if w := len([]rune(label)); w > maxW {
			maxW = w
		}
	}
	lines := make([]string, len(labels))
	for i := range labels {
		lines[i] = aboutAlignLabelValue(labels[i], values[i], maxW)
	}
	return lines
}

func (c Catalog) AboutTagline() string {
	if c.Lang == EN {
		return "Tracks your plants and reminds you when to water them."
	}
	return "Следит за вашими растениями и подсказывает, когда их поливать."
}

func (c Catalog) AboutDynamicHeader() string {
	if c.Lang == EN {
		return "Dynamic calculation (outdoor plants):"
	}
	return "Динамический расчет (для открытого грунта):"
}

func (c Catalog) AboutSeasonBullet() string {
	if c.Lang == EN {
		return "• Seasonality:  more watering in summer, less in winter (dormancy)"
	}
	return "• Сезонность:  летом — полив чаще, зимой — реже (анабиоз)"
}

func (c Catalog) AboutTemperatureBullet() string {
	if c.Lang == EN {
		return "• Temperature: heat speeds drying, cool weather slows it"
	}
	return "• Температура: жара ускоряет высыхание, прохлада замедляет"
}

func (c Catalog) AboutRainBullet() string {
	if c.Lang == EN {
		return "• Rain:        1–10 mm — watering later, >10 mm — already watered"
	}
	return "• Осадки:      1–10 мм — полив позже, >10 мм — уже полито"
}

func (c Catalog) AboutInfoLines() []string {
	lines := []string{
		version.Label(),
		c.SplashPlantStatusBoard(),
		"",
		c.AboutTagline(),
		"",
		c.AboutDynamicHeader(),
		c.AboutSeasonBullet(),
		c.AboutTemperatureBullet(),
		c.AboutRainBullet(),
		"",
	}
	return append(lines, c.aboutContactLines()...)
}

func (c Catalog) DetailsTitle() string {
	if c.Lang == EN {
		return "📋 DETAILS"
	}
	return "📋 ДЕТАЛИ"
}

func (c Catalog) DetailsTitlePlant(name, typeLabel string) string {
	if c.Lang == EN {
		return fmt.Sprintf("📋 DETAILS: %s [%s]", name, typeLabel)
	}
	return fmt.Sprintf("📋 ДЕТАЛИ: %s [%s]", name, typeLabel)
}

func (c Catalog) NoPlants() string {
	if c.Lang == EN {
		return "No plants yet."
	}
	return "Растений пока нет."
}

func (c Catalog) SelectPlant() string {
	if c.Lang == EN {
		return "Select a plant from the list."
	}
	return "Выберите растение в списке."
}

func (c Catalog) CalendarMonthTitle(year int, month time.Month) string {
	if c.Lang == EN {
		return fmt.Sprintf("📅 %s %d", month.String(), year)
	}
	ruMonths := []string{
		"", "Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
		"Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
	}
	m := int(month)
	name := month.String()
	if m >= 1 && m <= 12 {
		name = ruMonths[m]
	}
	return fmt.Sprintf("📅 %s %d", name, year)
}

func (c Catalog) CalendarWeekdays() []string {
	if c.Lang == EN {
		return []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	}
	return []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
}

func (c Catalog) CalendarSelectedDayTitle(day time.Time) string {
	if c.Lang == EN {
		return c.FormatDate(day) + ":"
	}
	return c.FormatDate(day) + ":"
}

func (c Catalog) CalendarNoTasks() string {
	if c.Lang == EN {
		return "No tasks for this day."
	}
	return "Нет задач на этот день."
}

func (c Catalog) CalendarTaskWater(name string) string {
	if c.Lang == EN {
		return "Water " + name
	}
	return "Полить " + name
}

func (c Catalog) CalendarTaskCompleted(name string) string {
	if c.Lang == EN {
		return "✓ Task completed: " + name
	}
	return "✓ Задача выполнена: " + name
}

func (c Catalog) CalendarTaskUnchecked(name string) string {
	if c.Lang == EN {
		return "Task unchecked: " + name
	}
	return "Задача снята: " + name
}

func (c Catalog) CalendarTaskAlreadyWatered(name string) string {
	if c.Lang == EN {
		return "Already watered today: " + name
	}
	return "Уже полито сегодня: " + name
}

func (c Catalog) LogTaskCompleted(name string) string {
	if c.Lang == EN {
		return "Task completed: water " + name
	}
	return "Задача выполнена: полить " + name
}

func (c Catalog) FooterCalendarHints() []FooterHint {
	tab := c.footerTabSwitchHint()
	if c.Lang == EN {
		return []FooterHint{
			tab,
			{Key: "⌨ h/j/k/l", Label: "day"},
			{Key: "Enter/Space", Label: "tasks"},
			{Key: "Space", Label: "toggle"},
			{Key: "Esc", Label: "grid"},
			{Key: "q", Label: "quit"},
		}
	}
	return []FooterHint{
		tab,
		{Key: "⌨ h/j/k/l", Label: "день"},
		{Key: "Enter/Space", Label: "задачи"},
		{Key: "Space", Label: "готово"},
		{Key: "Esc", Label: "сетка"},
		{Key: "q", Label: "выход"},
	}
}

func (c Catalog) FooterCareLogHints() []FooterHint {
	tab := c.footerTabSwitchHint()
	if c.Lang == EN {
		return []FooterHint{
			tab,
			{Key: "⌨ j/k", Label: "scroll"},
			{Key: "PgUp/PgDn", Label: "page"},
			{Key: "q", Label: "quit"},
		}
	}
	return []FooterHint{
		tab,
		{Key: "⌨ j/k", Label: "прокрутка"},
		{Key: "PgUp/PgDn", Label: "страница"},
		{Key: "q", Label: "выход"},
	}
}

func (c Catalog) CareLogScrollRange(from, to, total int) string {
	if c.Lang == EN {
		return fmt.Sprintf("%d–%d / %d", from, to, total)
	}
	return fmt.Sprintf("%d–%d / %d", from, to, total)
}

func (c Catalog) CareLogTitle() string {
	if c.Lang == EN {
		return "📝 Care Log"
	}
	return "📝 Журнал ухода"
}

func (c Catalog) CareLogColDate() string {
	if c.Lang == EN {
		return "Date/Time"
	}
	return "Дата/время"
}

func (c Catalog) CareLogColPlant() string {
	if c.Lang == EN {
		return "Plant"
	}
	return "Растение"
}

func (c Catalog) CareLogColEvent() string {
	if c.Lang == EN {
		return "Event"
	}
	return "Событие"
}

func (c Catalog) NoCareEvents() string {
	if c.Lang == EN {
		return "No care events yet."
	}
	return "Событий пока нет."
}

func (c Catalog) SettingsTitle() string {
	if c.Lang == EN {
		return "⚙ Settings"
	}
	return "⚙ Настройки"
}

func (c Catalog) SettingsCityNotSelected() string {
	if c.Lang == EN {
		return "not selected"
	}
	return "не выбран"
}

func (c Catalog) SettingsCityInputPlaceholder() string {
	if c.Lang == EN {
		return "type city name..."
	}
	return "введите название города..."
}

func (c Catalog) SettingsCityResults() string {
	if c.Lang == EN {
		return "Suggestions:"
	}
	return "Подсказки:"
}

func (c Catalog) SettingsCitySaved(city string) string {
	if c.Lang == EN {
		return "City selected: " + city
	}
	return "Город выбран: " + city
}

func (c Catalog) SettingsCitySearchFailed() string {
	if c.Lang == EN {
		return "City search failed. Check internet connection."
	}
	return "Не удалось получить список городов. Проверьте интернет."
}

func (c Catalog) SettingsOptionLanguage() string {
	if c.Lang == EN {
		return "Language"
	}
	return "Язык"
}

func (c Catalog) SettingsOptionCity() string {
	if c.Lang == EN {
		return "City"
	}
	return "Город"
}

func (c Catalog) SettingsOptionWeatherRefresh() string {
	if c.Lang == EN {
		return "Weather refresh"
	}
	return "Обновление погоды"
}

func (c Catalog) SettingsOptionAutoBackup() string {
	if c.Lang == EN {
		return "Auto-backup"
	}
	return "Автобэкап"
}

func (c Catalog) SettingsOptionTransparent() string {
	if c.Lang == EN {
		return "Transparent mode"
	}
	return "Прозрачный режим"
}

func (c Catalog) SettingsOptionFastBoot() string {
	if c.Lang == EN {
		return "Fastboot"
	}
	return "Быстрое включение"
}

func (c Catalog) SettingsOptionDefaultInterval() string {
	if c.Lang == EN {
		return "Default watering interval"
	}
	return "Интервал полива по умолчанию"
}

func (c Catalog) SettingsOptionFallbackTemp() string {
	if c.Lang == EN {
		return "Fallback temperature"
	}
	return "Стандартная температура"
}

func (c Catalog) SettingsOptionDatabase() string {
	if c.Lang == EN {
		return "Database"
	}
	return "База данных"
}

func (c Catalog) SettingsOptionBackup() string {
	if c.Lang == EN {
		return "Backup"
	}
	return "Бэкап"
}

func (c Catalog) SettingsOptionConfig() string {
	if c.Lang == EN {
		return "Config"
	}
	return "Конфиг"
}

func (c Catalog) SettingsOptionConfigBackup() string {
	if c.Lang == EN {
		return "Config backup"
	}
	return "Бэкап конфига"
}

func (c Catalog) SettingsDescLanguage() string {
	if c.Lang == EN {
		return "Interface language: Russian or English."
	}
	return "Язык интерфейса: русский или English."
}

func (c Catalog) SettingsDescCity() string {
	if c.Lang == EN {
		return "City for weather (Open-Meteo). Used to adjust outdoor watering from rain and temperature."
	}
	return "Город для погоды (Open-Meteo). Нужен для коррекции полива уличных растений по дождю и температуре."
}

func (c Catalog) SettingsDescWeatherRefresh() string {
	if c.Lang == EN {
		return "How often to fetch weather from the API (minutes). Use ←/→ to adjust by 1; you can also edit settings.toml."
	}
	return "Как часто запрашивать погоду по API (минуты). ←/→ меняют на 1; можно править settings.toml."
}

func (c Catalog) SettingsDescDefaultInterval() string {
	if c.Lang == EN {
		return "Default watering cycle in days for newly added plants."
	}
	return "Интервал полива по умолчанию в днях для новых растений."
}

func (c Catalog) SettingsDescFallbackTemp() string {
	if c.Lang == EN {
		return "Air temperature (°C) for drying calculations when live weather is unavailable."
	}
	return "Температура воздуха (°C) для расчёта увядания, когда нет данных погоды."
}

func (c Catalog) SettingsDescAutoBackup() string {
	if c.Lang == EN {
		return "Copy the database and settings.toml to backup files on each startup."
	}
	return "Копировать базу и settings.toml в файлы бэкапа при каждом запуске."
}

func (c Catalog) SettingsDescTransparent() string {
	if c.Lang == EN {
		return "Disable UI backgrounds so the terminal wallpaper shows through (handy in tile WMs)."
	}
	return "Отключает фоновые цвета интерфейса, чтобы фон терминала просвечивался (удобно в тайловых WM)."
}

func (c Catalog) SettingsDescFastBoot() string {
	if c.Lang == EN {
		return "Skip the splash screen and open the plant dashboard immediately on startup."
	}
	return "Пропускать заставку и сразу открывать карточки растений при старте."
}

func (c Catalog) SettingsWeatherRefreshValue(minutes int) string {
	if c.Lang == EN {
		return fmt.Sprintf("%d min", minutes)
	}
	return fmt.Sprintf("%d мин", minutes)
}

func (c Catalog) SettingsWeatherRefreshChanged(minutes int) string {
	if c.Lang == EN {
		return fmt.Sprintf("Weather refresh interval: %d min", minutes)
	}
	return fmt.Sprintf("Интервал обновления погоды: %d мин", minutes)
}

func (c Catalog) SettingsDefaultIntervalValue(days int) string {
	if c.Lang == EN {
		return fmt.Sprintf("%d days", days)
	}
	return fmt.Sprintf("%d дн.", days)
}

func (c Catalog) SettingsDefaultIntervalChanged(days int) string {
	if c.Lang == EN {
		return fmt.Sprintf("Default watering interval: %d days", days)
	}
	return fmt.Sprintf("Интервал полива по умолчанию: %d дн.", days)
}

func (c Catalog) SettingsFallbackTempValue(tempC float64) string {
	return fmt.Sprintf("%.1f°C", tempC)
}

func (c Catalog) SettingsFallbackTempChanged(tempC float64) string {
	if c.Lang == EN {
		return fmt.Sprintf("Fallback temperature: %.1f°C", tempC)
	}
	return fmt.Sprintf("Стандартная температура: %.1f°C", tempC)
}

func (c Catalog) SettingsTransparentEnabled() string {
	if c.Lang == EN {
		return "Transparent mode enabled."
	}
	return "Прозрачный режим включён."
}

func (c Catalog) SettingsTransparentDisabled() string {
	if c.Lang == EN {
		return "Transparent mode disabled."
	}
	return "Прозрачный режим выключен."
}

func (c Catalog) SettingsFastBootEnabled() string {
	if c.Lang == EN {
		return "Fastboot enabled (next start opens plants immediately)."
	}
	return "Быстрое включение включено (со следующего запуска сразу карточки растений)."
}

func (c Catalog) SettingsFastBootDisabled() string {
	if c.Lang == EN {
		return "Fastboot disabled (splash and animations on startup)."
	}
	return "Быстрое включение выключено (при старте будет сплэш и анимация)."
}

func (c Catalog) SettingsYes() string {
	if c.Lang == EN {
		return "Yes"
	}
	return "Да"
}

func (c Catalog) SettingsNo() string {
	if c.Lang == EN {
		return "No"
	}
	return "Нет"
}

func (c Catalog) SettingsBackupDone(dir string) string {
	if c.Lang == EN {
		return fmt.Sprintf("Backup saved to %s (data.db.bak, settings.toml.bak)", dir)
	}
	return fmt.Sprintf("Бэкап сохранён в %s (data.db.bak, settings.toml.bak)", dir)
}

func (c Catalog) SettingsBackupCancelled() string {
	if c.Lang == EN {
		return "Backup cancelled."
	}
	return "Бэкап отменён."
}

func (c Catalog) BackupConfirmTitle() string {
	if c.Lang == EN {
		return "💾 Backup data?"
	}
	return "💾 Сделать бэкап данных?"
}

func (c Catalog) BackupConfirmBody() string {
	if c.Lang == EN {
		return "The database and settings.toml will be copied to backup files (overwritten)."
	}
	return "База и settings.toml будут скопированы в файлы бэкапа (с перезаписью)."
}

func (c Catalog) BackupConfirmHint() string {
	if c.Lang == EN {
		return "← →: select  •  Enter: confirm  •  y: Yes  •  n / Esc: No"
	}
	return "← →: выбор  •  Enter: подтвердить  •  y: Да  •  n / Esc: Нет"
}

func (c Catalog) SettingsAutoBackupEnabled() string {
	if c.Lang == EN {
		return "Auto-backup enabled (runs on startup)."
	}
	return "Автобэкап включён (при каждом запуске)."
}

func (c Catalog) SettingsAutoBackupDisabled() string {
	if c.Lang == EN {
		return "Auto-backup disabled."
	}
	return "Автобэкап выключен."
}

func (c Catalog) FooterSettingsHints() []FooterHint {
	tab := c.footerTabSwitchHint()
	if c.Lang == EN {
		return []FooterHint{
			tab,
			{Key: "⌨ j/k", Label: "navigate"},
			{Key: "h/l", Label: "change value"},
			{Key: "Space", Label: "toggle"},
			{Key: "b", Label: "backup now"},
			{Key: "q", Label: "quit"},
		}
	}
	return []FooterHint{
		tab,
		{Key: "⌨ j/k", Label: "навигация"},
		{Key: "h/l", Label: "изменить значение"},
		{Key: "Space", Label: "переключить"},
		{Key: "b", Label: "сделать бэкап"},
		{Key: "q", Label: "выход"},
	}
}

func (c Catalog) LanguageChanged() string {
	if c.Lang == EN {
		return "Language set to English"
	}
	return "Язык переключён на русский"
}

func (c Catalog) FooterHints() []FooterHint {
	tab := c.footerTabSwitchHint()
	if c.Lang == EN {
		return []FooterHint{
			tab,
			{Key: "⌨ h/j/k/l", Label: "navigate"},
			{Key: "a", Label: "add"},
			{Key: "e", Label: "edit"},
			{Key: "w", Label: "water"},
			{Key: "s", Label: "postpone"},
			{Key: "d", Label: "delete"},
			{Key: "q", Label: "quit"},
		}
	}
	return []FooterHint{
		tab,
		{Key: "⌨ h/j/k/l", Label: "навигация"},
		{Key: "a", Label: "добавить"},
		{Key: "e", Label: "изменить"},
		{Key: "w", Label: "полить"},
		{Key: "s", Label: "сдвинуть"},
		{Key: "d", Label: "удалить"},
		{Key: "q", Label: "выход"},
	}
}

func (c Catalog) FooterOverlayHints() []FooterHint {
	if c.Lang == EN {
		return []FooterHint{{Key: "Esc", Label: "cancel"}}
	}
	return []FooterHint{{Key: "Esc", Label: "отмена"}}
}

func (c Catalog) FooterBackupConfirmHints() []FooterHint {
	if c.Lang == EN {
		return []FooterHint{
			{Key: "← →", Label: "select"},
			{Key: "Enter/y", Label: "yes"},
			{Key: "n/Esc", Label: "no"},
		}
	}
	return []FooterHint{
		{Key: "← →", Label: "выбор"},
		{Key: "Enter/y", Label: "да"},
		{Key: "n/Esc", Label: "нет"},
	}
}

func (c Catalog) DeleteConfirmTitle(name string) string {
	if c.Lang == EN {
		return fmt.Sprintf("⚠  Delete plant \"%s\"?", name)
	}
	return fmt.Sprintf("⚠  Удалить растение «%s»?", name)
}

func (c Catalog) DeleteConfirmBody() string {
	if c.Lang == EN {
		return "This will remove the plant and its care log history."
	}
	return "Растение и вся история ухода будут удалены безвозвратно."
}

func (c Catalog) DeleteConfirmHint() string {
	if c.Lang == EN {
		return "y: Yes, delete  •  n / Esc: Cancel"
	}
	return "y: Да, удалить  •  n / Esc: Отмена"
}

func (c Catalog) AddPlantTitle() string {
	if c.Lang == EN {
		return "➕ ADD PLANT"
	}
	return "➕ НОВОЕ РАСТЕНИЕ"
}

func (c Catalog) EditPlantTitle(name string) string {
	if c.Lang == EN {
		return fmt.Sprintf("✎ EDIT PLANT — %s", name)
	}
	return fmt.Sprintf("✎ ИЗМЕНИТЬ — %s", name)
}

func (c Catalog) FieldName() string {
	if c.Lang == EN {
		return "Name"
	}
	return "Название"
}

func (c Catalog) FieldLocation() string {
	if c.Lang == EN {
		return "Location"
	}
	return "Локация"
}

func (c Catalog) FieldInterval() string {
	if c.Lang == EN {
		return "Water every (days)"
	}
	return "Полив каждые (дней)"
}

func (c Catalog) FieldOutdoor() string {
	if c.Lang == EN {
		return "Outdoor"
	}
	return "Улица"
}

func (c Catalog) Yes() string {
	if c.Lang == EN {
		return "Yes"
	}
	return "Да"
}

func (c Catalog) No() string {
	if c.Lang == EN {
		return "No"
	}
	return "Нет"
}

func (c Catalog) AddPlantHint() string {
	if c.Lang == EN {
		return "⌨Tab: next field  •  Space: toggle outdoor  •  Enter: save  •  Esc: cancel"
	}
	return "⌨ Tab: следующее поле  •  Space: улица вкл/выкл  •  Enter: сохранить  •  Esc: отмена"
}

func (c Catalog) StatusPlantAdded(name string) string {
	if c.Lang == EN {
		return fmt.Sprintf("Plant added: %s", name)
	}
	return fmt.Sprintf("Растение добавлено: %s", name)
}

func (c Catalog) StatusPlantDeleted(name string) string {
	if c.Lang == EN {
		return fmt.Sprintf("Plant deleted: %s", name)
	}
	return fmt.Sprintf("Растение удалено: %s", name)
}

func (c Catalog) StatusDeleteCancelled() string {
	if c.Lang == EN {
		return "Deletion cancelled"
	}
	return "Удаление отменено"
}

func (c Catalog) StatusAddCancelled() string {
	if c.Lang == EN {
		return "Add plant cancelled"
	}
	return "Добавление отменено"
}

func (c Catalog) StatusEditCancelled() string {
	if c.Lang == EN {
		return "Edit cancelled"
	}
	return "Изменение отменено"
}

func (c Catalog) StatusPlantUpdated(name string) string {
	if c.Lang == EN {
		return fmt.Sprintf("Plant updated: %s", name)
	}
	return fmt.Sprintf("Растение обновлено: %s", name)
}

func (c Catalog) StatusNoPlantSelected() string {
	if c.Lang == EN {
		return "No plant selected"
	}
	return "Растение не выбрано"
}

func (c Catalog) StatusNameRequired() string {
	if c.Lang == EN {
		return "Name is required"
	}
	return "Укажите название"
}

func (c Catalog) StatusLocationRequired() string {
	if c.Lang == EN {
		return "Location is required"
	}
	return "Укажите локацию"
}

func (c Catalog) StatusInvalidInterval() string {
	if c.Lang == EN {
		return "Interval must be 1–365 days"
	}
	return "Интервал должен быть от 1 до 365 дней"
}

func (c Catalog) LogPlantAdded() string {
	if c.Lang == EN {
		return "Plant added to greenhouse"
	}
	return "Растение добавлено в оранжерею"
}

func (c Catalog) LogPlantUpdated(days int) string {
	if c.Lang == EN {
		return fmt.Sprintf("Plant settings updated (cycle: %d days)", days)
	}
	return fmt.Sprintf("Настройки растения обновлены (цикл: %d дн.)", days)
}

func (c Catalog) Location(stored string) string {
	switch stored {
	case "living_room", "Гостиная", "Living Room":
		if c.Lang == EN {
			return "Living Room"
		}
		return "Гостиная"
	case "terrace", "Терраса", "Terrace":
		if c.Lang == EN {
			return "Terrace"
		}
		return "Терраса"
	case "office", "Кабинет", "Office":
		if c.Lang == EN {
			return "Office"
		}
		return "Кабинет"
	default:
		return stored
	}
}

func (c Catalog) PlantTypeLabel(p model.Plant) string {
	loc := c.Location(p.Location)
	if p.IsOutdoor {
		if c.Lang == EN {
			return fmt.Sprintf("Outdoor (%s)", loc)
		}
		return fmt.Sprintf("Улица (%s)", loc)
	}
	if c.Lang == EN {
		return fmt.Sprintf("Indoor (%s)", loc)
	}
	return fmt.Sprintf("Комната (%s)", loc)
}

func (c Catalog) RowStatusWaterNow() string {
	if c.Lang == EN {
		return "WATER TODAY!"
	}
	return "ПОЛИТЬ СЕГОДНЯ!"
}

func (c Catalog) RowStatusShiftedRain(date time.Time) string {
	if c.Lang == EN {
		return "Water: " + c.FormatDate(date) + " (rain)"
	}
	return "Полив: " + c.FormatDate(date) + " (дождь)"
}

func (c Catalog) RowStatusWatering(date time.Time) string {
	if c.Lang == EN {
		return "Water: " + c.FormatDate(date)
	}
	return "Полив: " + c.FormatDate(date)
}

func (c Catalog) CLIHeaderPlant() string {
	if c.Lang == EN {
		return "PLANT"
	}
	return "РАСТЕНИЕ"
}

func (c Catalog) CLIHeaderStatus() string {
	if c.Lang == EN {
		return "STATUS"
	}
	return "СТАТУС"
}

func (c Catalog) CLIHeaderWatering() string {
	if c.Lang == EN {
		return "WATERING"
	}
	return "ПОЛИВ"
}

func (c Catalog) CLIPlantName(name string, outdoor bool) string {
	if outdoor {
		if c.Lang == EN {
			return fmt.Sprintf("%s (outdoor)", name)
		}
		return fmt.Sprintf("%s (улица)", name)
	}
	if c.Lang == EN {
		return fmt.Sprintf("%s (indoor)", name)
	}
	return fmt.Sprintf("%s (комната)", name)
}

func (c Catalog) CLIWateringUrgent() string {
	if c.Lang == EN {
		return "URGENT!"
	}
	return "СРОЧНО!"
}

func (c Catalog) CLIWateringAlready() string {
	if c.Lang == EN {
		return "Already watered"
	}
	return "Уже полито"
}

func (c Catalog) CLIWateringInDays(days int) string {
	if days < 1 {
		days = 1
	}
	if c.Lang == EN {
		if days == 1 {
			return "In 1 day"
		}
		return fmt.Sprintf("In %d days", days)
	}
	n := days % 100
	switch {
	case n%10 == 1 && n != 11:
		return fmt.Sprintf("Через %d день", days)
	case n%10 >= 2 && n%10 <= 4 && (n < 12 || n > 14):
		return fmt.Sprintf("Через %d дня", days)
	default:
		return fmt.Sprintf("Через %d дней", days)
	}
}

func (c Catalog) CLINoPlants() string {
	if c.Lang == EN {
		return "No plants yet."
	}
	return "Растений пока нет."
}

func (c Catalog) CLISummaryOK() string {
	return "🌿 kiri: ok"
}

func (c Catalog) CLISummaryNeedsWater(count int) string {
	if count < 1 {
		count = 1
	}
	if c.Lang == EN {
		if count == 1 {
			return fmt.Sprintf("💧 kiri: %d needs watering!", count)
		}
		return fmt.Sprintf("💧 kiri: %d need watering!", count)
	}
	n := count % 100
	verb := "требуют"
	if n%10 == 1 && n != 11 {
		verb = "требует"
	}
	return fmt.Sprintf("💧 kiri: %d %s полива!", count, verb)
}

func (c Catalog) DetailBaseCycle(days int) string {
	if c.Lang == EN {
		return fmt.Sprintf("Base cycle: every %d days.", days)
	}
	return fmt.Sprintf("Базовый цикл: каждые %d дней.", days)
}

func (c Catalog) DetailStatusRain(date time.Time) string {
	if c.Lang == EN {
		return fmt.Sprintf(
			"Engine moved watering to %s because rain was recorded outside.",
			c.FormatDate(date),
		)
	}
	return fmt.Sprintf(
		"Движок сдвинул полив на %s, так как на улице зафиксирован дождь.",
		c.FormatDate(date),
	)
}

func (c Catalog) DetailStatusCritical() string {
	if c.Lang == EN {
		return "Critical thirst — immediate watering required."
	}
	return "Критическая жажда — требуется немедленный полив."
}

func (c Catalog) DetailStatusScheduled(date time.Time) string {
	if c.Lang == EN {
		return fmt.Sprintf("Next watering scheduled for %s.", c.FormatDate(date))
	}
	return fmt.Sprintf("Следующий полив запланирован на %s.", c.FormatDate(date))
}

func (c Catalog) DetailStatusPrefix() string {
	if c.Lang == EN {
		return "Status: "
	}
	return "Статус: "
}

func (c Catalog) DetailLastLogPrefix() string {
	if c.Lang == EN {
		return "Last log: "
	}
	return "Последний лог: "
}

func (c Catalog) NoLogEntries() string {
	if c.Lang == EN {
		return "No entries yet."
	}
	return "Записей пока нет."
}

func (c Catalog) PostponeTip(count int) string {
	if c.Lang == EN {
		return fmt.Sprintf(
			"Tip: %d postponements in a row — consider increasing base cycle by +1 day.",
			count,
		)
	}
	return fmt.Sprintf(
		"Подсказка: %d отложения подряд — увеличьте базовый цикл на +1 день.",
		count,
	)
}

func (c Catalog) LogWateredNormal() string {
	if c.Lang == EN {
		return "Normal watering (2L water added)"
	}
	return "Полив нормальный (внесено 2л воды)"
}

func (c Catalog) LogPostponed() string {
	if c.Lang == EN {
		return "Watering postponed (+1 day)"
	}
	return "Полив отложен (+1 day)"
}

func (c Catalog) LogWeatherConfirmed() string {
	if c.Lang == EN {
		return "System/weather shift confirmed"
	}
	return "Подтверждён системный/погодный сдвиг"
}

func (c Catalog) LogRainWatered(precipMM float64) string {
	if c.Lang == EN {
		return fmt.Sprintf("Watered by rain — heavy precipitation (%.1fmm)", precipMM)
	}
	return fmt.Sprintf("Полит дождём — сильные осадки (%.1fмм)", precipMM)
}

func (c Catalog) LogRainShifted(precipMM, shiftHours, waterAdded float64) string {
	if c.Lang == EN {
		return fmt.Sprintf("Postponed due to rain (%.1fmm): +%.0f%% water, %.1fh shift", precipMM, waterAdded, shiftHours)
	}
	return fmt.Sprintf("Перенесён из-за дождя (%.1fмм): +%.0f%% воды, сдвиг %.1f ч", precipMM, waterAdded, shiftHours)
}

func (c Catalog) StatusRainApplied(count int, precipMM float64) string {
	if c.Lang == EN {
		return fmt.Sprintf("Weather adjustment applied for %d outdoor plants (%.1fmm)", count, precipMM)
	}
	return fmt.Sprintf("Погодная корректировка применена к %d уличным растениям (%.1fмм)", count, precipMM)
}

func (c Catalog) LogIntervalSuggestion(days int) string {
	if c.Lang == EN {
		return fmt.Sprintf("Recommendation: increase cycle to %d days", days)
	}
	return fmt.Sprintf("Рекомендация: увеличить цикл до %d дней", days)
}

func (c Catalog) StatusWatered(name string) string {
	if c.Lang == EN {
		return fmt.Sprintf("Watered: %s — tank at 100%%", name)
	}
	return fmt.Sprintf("Полито: %s — бак 100%%", name)
}

func (c Catalog) StatusPostponed(name string) string {
	if c.Lang == EN {
		return fmt.Sprintf("Watering postponed: %s", name)
	}
	return fmt.Sprintf("Полив отложен: %s", name)
}

func (c Catalog) TranslateLogMessage(msg string) string {
	known := map[string][2]string{
		"Полив нормальный (внесено 2л воды)":   {"Normal watering (2L water added)", "Полив нормальный (внесено 2л воды)"},
		"Жажда стабильна на 80%":               {"Thirst stable at 80%", "Жажда стабильна на 80%"},
		"Полит дождём — сильные осадки (12мм)": {"Watered by rain — heavy precipitation (12mm)", "Полит дождём — сильные осадки (12мм)"},
		"Перенесён из-за дождя: +20% воды":     {"Postponed due to rain: +20% water", "Перенесён из-за дождя: +20% воды"},
		"Критическая жажда — ПОЛИТЬ СЕГОДНЯ!":  {"Critical thirst — WATER TODAY!", "Критическая жажда — ПОЛИТЬ СЕГОДНЯ!"},
		"Полив отложен (+1 день)":              {"Watering postponed (+1 day)", "Полив отложен (+1 день)"},
		"Полив отложен повторно (+1 день)":     {"Watering postponed again (+1 day)", "Полив отложен повторно (+1 день)"},
		"Normal watering (2L water added)":     {"Normal watering (2L water added)", "Полив нормальный (внесено 2л воды)"},
		"Watering postponed (+1 day)":          {"Watering postponed (+1 day)", "Полив отложен (+1 день)"},
		"Watering postponed again (+1 day)":    {"Watering postponed again (+1 day)", "Полив отложен повторно (+1 день)"},
		"Plant added to greenhouse":            {"Plant added to greenhouse", "Растение добавлено в оранжерею"},
		"Plant removed from greenhouse":        {"Plant removed from greenhouse", "Растение удалено из оранжереи"},
		"Растение добавлено в оранжерею":       {"Plant added to greenhouse", "Растение добавлено в оранжерею"},
		"Растение удалено из оранжереи":        {"Plant removed from greenhouse", "Растение удалено из оранжереи"},
	}
	if pair, ok := known[msg]; ok {
		if c.Lang == EN {
			return pair[0]
		}
		return pair[1]
	}
	return msg
}

func (c Catalog) FormatDate(t time.Time) string {
	if c.Lang == EN {
		return t.Format("January 2")
	}
	ruMonths := []string{
		"", "января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	m := int(t.Month())
	if m < 1 || m > 12 {
		return t.Format("02.01.2006")
	}
	return fmt.Sprintf("%d %s", t.Day(), ruMonths[m])
}

func (c Catalog) FormatDateTime(t time.Time) string {
	if c.Lang == EN {
		enMonths := []string{
			"", "Jan", "Feb", "Mar", "Apr", "May", "Jun",
			"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
		}
		m := int(t.Month())
		if m < 1 || m > 12 {
			return t.Format("2006-01-02 15:04")
		}
		return fmt.Sprintf("%s %d %02d:%02d", enMonths[m], t.Day(), t.Hour(), t.Minute())
	}
	ruMonths := []string{
		"", "янв", "фев", "мар", "апр", "май", "июн",
		"июл", "авг", "сен", "окт", "ноя", "дек",
	}
	m := int(t.Month())
	if m < 1 || m > 12 {
		return t.Format("02.01.2006 15:04")
	}
	return fmt.Sprintf("%d %s %02d:%02d", t.Day(), ruMonths[m], t.Hour(), t.Minute())
}

func (c Catalog) FormatLogLine(date time.Time, message string) string {
	return fmt.Sprintf("%s — %s", c.FormatDate(date), c.TranslateLogMessage(message))
}
