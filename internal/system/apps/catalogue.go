package apps

// App describes a Flatpak application managed by lumina.
type App struct {
	Name   string
	FlatID string
}

// Catalogue lists all Flatpak apps available for installation.
// To add a new app, append a new App entry here.
var Catalogue = []App{
	{Name: "Loupe - Visualizador de Fotos", FlatID: "org.gnome.Loupe"},
	{Name: "Celluloid - Visualizador de Vídeo", FlatID: "io.github.celluloid_player.Celluloid"},
	{Name: "VLC - Visualizador de Vídeo", FlatID: "org.videolan.VLC"},
	{Name: "Vinyl - Player de Música", FlatID: "page.codeberg.M23Snezhok.Vinyl"},
	{Name: "Calculator - Calculadora", FlatID: "org.gnome.Calculator"},
	{Name: "Resources - Monitor do Sistema", FlatID: "net.nokyan.Resources"},
	{Name: "Gradia - Captura de Tela", FlatID: "be.alexandervanhee.gradia"},
	{Name: "Apostrophe - Markdown", FlatID: "org.gnome.gitlab.somas.Apostrophe"},
	{Name: "Eyedropper - Conta Gotas", FlatID: "com.github.finefindus.eyedropper"},
	{Name: "Gear Lever - AppImages", FlatID: "it.mijorus.gearlever"},
	{Name: "Web Apps", FlatID: "net.codelogistics.webapps"},
	{Name: "Flatseal - Gerenciar Flatpak", FlatID: "com.github.tchx84.Flatseal"},
	{Name: "Zen Browser", FlatID: "app.zen_browser.zen"},
	{Name: "Firefox", FlatID: "org.mozilla.firefox"},
	{Name: "Chromium", FlatID: "org.chromium.Chromium"},
	{Name: "FileZilla", FlatID: "org.filezillaproject.Filezilla"},
	{Name: "Inkscape", FlatID: "org.inkscape.Inkscape"},
	{Name: "Krita", FlatID: "org.kde.krita"},
	{Name: "Penpot", FlatID: "com.authormore.penpotdesktop"},
	{Name: "LibreOffice", FlatID: "org.libreoffice.LibreOffice"},
	{Name: "OnlyOffice", FlatID: "org.onlyoffice.desktopeditors"},
	{Name: "AnyDesk", FlatID: "com.anydesk.Anydesk"},
	{Name: "Meld - File Compare", FlatID: "org.gnome.meld"},
	{Name: "Minecraft", FlatID: "io.mrarm.mcpelauncher"},
	{Name: "Minecraft Java", FlatID: "org.prismlauncher.PrismLauncher"},
	{Name: "Ente Auth - Segurança", FlatID: "io.ente.auth"},
	{Name: "Font Downloader", FlatID: "org.gustavoperedo.FontDownloader"},
	{Name: "FreeTube - YouTube", FlatID: "io.freetubeapp.FreeTube"},
}
