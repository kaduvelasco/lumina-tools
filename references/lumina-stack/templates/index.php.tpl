<?php
/**
 * LuminaStack - Dashboard de Desenvolvimento
 * Local: /srv/workspace/www/html/index.php
 */

// Lê as versões PHP do .env da stack
$envFile = __DIR__ . "/../../docker/.env";
$phpVersions = [];

if (file_exists($envFile)) {
    $envContent = file_get_contents($envFile);
    // Usa \r?\n para capturar apenas a linha PHP_VERSIONS, sem vazar para linhas seguintes
    if (preg_match('/^PHP_VERSIONS=(.+)/m', $envContent, $matches)) {
        $phpVersions = preg_split('/\s+/', trim($matches[1]), -1, PREG_SPLIT_NO_EMPTY);
    }
}

// Fallback caso não consiga ler o .env (lista completa de versões suportadas)
if (empty($phpVersions)) {
    $phpVersions = ["7.4", "8.0", "8.1", "8.2", "8.3", "8.4"];
}
?>
<!DOCTYPE html>
<html lang="pt-br">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LuminaStack | Dev Panel</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
</head>
<body class="bg-slate-900 text-slate-200 font-sans min-h-screen flex flex-col items-center justify-center p-6">

    <div class="max-w-4xl w-full">

        <!-- Cabeçalho -->
        <div class="text-center mb-10">
            <h1 class="text-5xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-emerald-400 mb-2">
                LuminaStack
            </h1>
            <p class="text-slate-400 uppercase tracking-widest text-sm">Ambiente de Desenvolvimento Docker</p>
        </div>

        <!-- Cards de versões PHP -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <?php foreach ($phpVersions as $version): ?>
                <?php
                    $vClean = str_replace(".", "", $version);
                    $url    = "http://php{$vClean}.localhost/info.php";
                ?>
                <a href="<?php echo htmlspecialchars($url); ?>" target="_blank"
                   class="group block p-6 bg-slate-800 border border-slate-700 rounded-2xl hover:border-blue-500 transition-all duration-300 shadow-xl hover:shadow-blue-500/10">
                    <div class="flex items-center justify-between">
                        <div>
                            <span class="text-xs font-bold text-blue-400 uppercase tracking-tight">Versão Instalada</span>
                            <h2 class="text-3xl font-bold text-white mt-1">PHP <?php echo htmlspecialchars($version); ?></h2>
                        </div>
                        <div class="bg-slate-700 p-4 rounded-xl group-hover:bg-blue-600 transition-colors">
                            <i class="fa-brands fa-php text-3xl text-white"></i>
                        </div>
                    </div>
                    <div class="mt-4 flex items-center text-slate-400 text-sm">
                        <i class="fa-solid fa-link mr-2"></i>
                        <span>php<?php echo htmlspecialchars($vClean); ?>.localhost</span>
                    </div>
                </a>
            <?php endforeach; ?>
        </div>

        <!-- Informações do ambiente -->
        <div class="mt-12 grid grid-cols-1 md:grid-cols-4 gap-4 text-center">
            <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50">
                <i class="fa-solid fa-database text-emerald-400 mb-2 text-lg"></i>
                <p class="text-xs text-slate-500 uppercase mb-1">MariaDB</p>
                <p class="font-mono text-sm">localhost:3306</p>
            </div>
            <div class="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50">
                <i class="fa-solid fa-folder-open text-amber-400 mb-2 text-lg"></i>
                <p class="text-xs text-slate-500 uppercase mb-1">Projetos em</p>
                <p class="font-mono text-sm">/srv/workspace/www/html</p>
            </div>
        </div>

    </div>

</body>
</html>
