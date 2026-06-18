#!/usr/bin/env bash
# =============================================================================
# Script Name     : file-models.sh
# Description     : Create template files in the user templates directory
# Version         : 2.0.0
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

if [[ ! -f "$SCRIPT_DIR/../lib/utils.sh" ]]; then
    printf '\033[0;31m❌ Erro fatal: ../lib/utils.sh não encontrado.\033[0m\n' >&2
    exit 1
fi
# shellcheck source=/dev/null
source "$SCRIPT_DIR/../lib/utils.sh"

# --- globals ---
declare -a TEMPLATE_NAMES=()
declare -a TEMPLATE_FILES=()
declare -i MENU_INDEX=1
declare -a SELECTED_TEMPLATES=()

# --- dependencies ---
ensure_dependencies() {
    ensure_pkg python3
    ensure_pkg xdg-user-dirs
}

# --- template creators ---
create_docx() {
    local destination="$1"
    python3 - "$destination" <<'PYEOF'
import sys, zipfile
dest = sys.argv[1]
content_types = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>'
rels = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>'
document = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p/></w:body></w:document>'
with zipfile.ZipFile(dest, "w", zipfile.ZIP_DEFLATED) as z:
    z.writestr("[Content_Types].xml", content_types)
    z.writestr("_rels/.rels", rels)
    z.writestr("word/document.xml", document)
PYEOF
}

create_xlsx() {
    local destination="$1"
    python3 - "$destination" <<'PYEOF'
import sys, zipfile
dest = sys.argv[1]
content_types = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>'
rels = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>'
workbook = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Sheet1" sheetId="1" r:id="rId1"/></sheets></workbook>'
workbook_rels = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>'
sheet = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData/></worksheet>'
with zipfile.ZipFile(dest, "w", zipfile.ZIP_DEFLATED) as z:
    z.writestr("[Content_Types].xml", content_types)
    z.writestr("_rels/.rels", rels)
    z.writestr("xl/workbook.xml", workbook)
    z.writestr("xl/_rels/workbook.xml.rels", workbook_rels)
    z.writestr("xl/worksheets/sheet1.xml", sheet)
PYEOF
}

create_pptx() {
    local destination="$1"
    python3 - "$destination" <<'PYEOF'
import sys, zipfile
dest = sys.argv[1]
content_types = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/></Types>'
rels = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/></Relationships>'
presentation = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><p:sldMasterIdLst/><p:sldSz cx="9144000" cy="6858000"/><p:notesSz cx="6858000" cy="9144000"/></p:presentation>'
with zipfile.ZipFile(dest, "w", zipfile.ZIP_DEFLATED) as z:
    z.writestr("[Content_Types].xml", content_types)
    z.writestr("_rels/.rels", rels)
    z.writestr("ppt/presentation.xml", presentation)
PYEOF
}

create_odt() {
    local destination="$1"
    python3 - "$destination" <<'PYEOF'
import sys, zipfile
dest = sys.argv[1]
mimetype = 'application/vnd.oasis.opendocument.text'
manifest = '<?xml version="1.0" encoding="UTF-8"?><manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2"><manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.text"/><manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/></manifest:manifest>'
content = '<?xml version="1.0" encoding="UTF-8"?><office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" office:version="1.2"><office:body><office:text><text:p/></office:text></office:body></office:document-content>'
with zipfile.ZipFile(dest, "w") as z:
    info = zipfile.ZipInfo("mimetype")
    info.compress_type = zipfile.ZIP_STORED
    z.writestr(info, mimetype)
    z.writestr("META-INF/manifest.xml", manifest)
    z.writestr("content.xml", content)
PYEOF
}

create_ods() {
    local destination="$1"
    python3 - "$destination" <<'PYEOF'
import sys, zipfile
dest = sys.argv[1]
mimetype = 'application/vnd.oasis.opendocument.spreadsheet'
manifest = '<?xml version="1.0" encoding="UTF-8"?><manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2"><manifest:file-entry manifest:full-path="/" manifest:media-type="application/vnd.oasis.opendocument.spreadsheet"/><manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/></manifest:manifest>'
content = '<?xml version="1.0" encoding="UTF-8"?><office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" office:version="1.2"><office:body><office:spreadsheet><table:table table:name="Planilha1"><table:table-row><table:table-cell/></table:table-row></table:table></office:spreadsheet></office:body></office:document-content>'
with zipfile.ZipFile(dest, "w") as z:
    info = zipfile.ZipInfo("mimetype")
    info.compress_type = zipfile.ZIP_STORED
    z.writestr(info, mimetype)
    z.writestr("META-INF/manifest.xml", manifest)
    z.writestr("content.xml", content)
PYEOF
}

create_txt() {
    local destination="$1"
    printf 'Criado em: %s\n' "$(date '+%d/%m/%Y %H:%M')" > "${destination}"
}

create_html() {
    local destination="$1"
    cat <<'EOF' > "${destination}"
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Novo Projeto</title>
</head>
<body>
</body>
</html>
EOF
}

create_css() {
    local destination="$1"
    cat <<'EOF' > "${destination}"
/* ============================================================
   Stylesheet
   ============================================================ */

*, *::before, *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}
EOF
}

create_js() {
    local destination="$1"
    cat <<'EOF' > "${destination}"
'use strict';

EOF
}

create_py() {
    local destination="$1"
    cat <<'EOF' > "${destination}"
#!/usr/bin/env python3


def main():
    pass


if __name__ == '__main__':
    main()
EOF
}

create_php() {
    local destination="$1"
    cat <<'EOF' > "${destination}"
<?php

declare(strict_types=1);

EOF
}

create_sh() {
    local destination="$1"
    cat <<'EOF' > "${destination}"
#!/usr/bin/env bash
set -euo pipefail

EOF
}

# --- dispatch ---
create_template() {
    local filename="$1"
    local dest="$2"
    case "${filename}" in
        *.docx) create_docx "${dest}" ;;
        *.xlsx) create_xlsx "${dest}" ;;
        *.pptx) create_pptx "${dest}" ;;
        *.odt)  create_odt  "${dest}" ;;
        *.ods)  create_ods  "${dest}" ;;
        *.txt)  create_txt  "${dest}" ;;
        *.html) create_html "${dest}" ;;
        *.css)  create_css  "${dest}" ;;
        *.js)   create_js   "${dest}" ;;
        *.py)   create_py   "${dest}" ;;
        *.php)  create_php  "${dest}" ;;
        *.sh)   create_sh   "${dest}" ;;
        *)      warn "Tipo desconhecido: ${filename}" ;;
    esac
}

# --- menu helpers ---
append_menu_item() {
    local label="$1"
    local filename="$2"
    local index="$3"
    local tdir="$4"

    TEMPLATE_NAMES+=("${label}")
    TEMPLATE_FILES+=("${filename}")

    if [[ -f "${tdir}/${filename}" ]]; then
        printf '%b\n' "  ${C2}${index}.${RESET} ${SIM_WARN} ${label} ${C3}(já existe)${RESET}"
        return 0
    fi

    printf '%b\n' "  ${C2}${index}.${RESET} ${label}"
}

process_selections() {
    local -a raw_choices=("$@")
    local choice
    local selected_name

    SELECTED_TEMPLATES=()

    for choice in "${raw_choices[@]}"; do
        if [[ "${choice}" == '0' ]]; then
            return 10
        fi

        if [[ "${choice}" == 'all' ]]; then
            SELECTED_TEMPLATES=("${TEMPLATE_FILES[@]}")
            return 0
        fi

        if ! [[ "${choice}" =~ ^[0-9]+$ ]]; then
            warn "Entrada ignorada: ${choice}"
            continue
        fi

        if ((choice < 1 || choice > ${#TEMPLATE_NAMES[@]})); then
            warn "Opção fora do intervalo: ${choice}"
            continue
        fi

        selected_name="${TEMPLATE_NAMES[$((choice - 1))]}"
        info "Selecionado: ${selected_name}"
        SELECTED_TEMPLATES+=("${TEMPLATE_FILES[$((choice - 1))]}")
    done
}

execute_templates() {
    local templates_dir="$1"
    local -a created=() skipped=()
    local filename dest

    if ((${#SELECTED_TEMPLATES[@]} == 0)); then
        warn "Nenhum modelo foi selecionado."
        return 0
    fi

    for filename in "${SELECTED_TEMPLATES[@]}"; do
        dest="${templates_dir}/${filename}"
        if [[ -f "${dest}" ]]; then
            warn "Pulado (já existe): ${filename}"
            skipped+=("${filename}")
            continue
        fi
        create_template "${filename}" "${dest}"
        if [[ "${filename}" == *.sh || "${filename}" == *.py ]]; then
            chmod +x -- "${dest}"
        else
            chmod 644 -- "${dest}"
        fi
        success "Criado: ${filename}"
        created+=("${filename}")
    done

    printf '%b\n' ""
    if ((${#created[@]} > 0)); then
        success "${#created[@]} modelo(s) criado(s) com sucesso."
    fi
    if ((${#skipped[@]} > 0)); then
        warn "${#skipped[@]} modelo(s) já existiam e foram mantidos."
    fi
}

# --- interface ---
show_header() {
    show_lumina_header
}

show_menu() {
    local -a selections
    local process_status=0
    local templates_dir

    show_header
    ensure_dependencies

    templates_dir="$(xdg-user-dir TEMPLATES 2>/dev/null || printf '%s' "${HOME}/Templates")"
    mkdir -p -- "${templates_dir}"

    TEMPLATE_NAMES=()
    TEMPLATE_FILES=()
    MENU_INDEX=1

    printf '%b\n' "Selecione os modelos pelo número ou use ${C2}all${RESET}."
    printf '%b\n' "${C3}${SIM_WARN} Modelos já existentes serão mantidos.${RESET}"
    printf '%b\n' ""
    append_menu_item "Documento Word (.docx)"          "Documento.docx"    "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "Planilha Excel (.xlsx)"           "Planilha.xlsx"     "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "Apresentação PowerPoint (.pptx)"  "Apresentacao.pptx" "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "Documento LibreOffice (.odt)"     "Documento.odt"     "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "Planilha LibreOffice (.ods)"      "Planilha.ods"      "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "Texto (.txt)"                     "Texto.txt"         "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "HTML (.html)"                     "HTML.html"         "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "CSS (.css)"                       "Estilo.css"        "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "JavaScript (.js)"                 "Script.js"         "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "Python (.py)"                     "Script.py"         "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "PHP (.php)"                       "PHP.php"           "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))
    append_menu_item "Shell (.sh)"                      "Shell.sh"          "${MENU_INDEX}" "${templates_dir}"; MENU_INDEX=$((MENU_INDEX + 1))

    printf '%b\n' ""
    printf '%b\n' "  ${C2}all${RESET} Criar todos"
    printf '%b\n' "  ${C1}0.${RESET} Voltar"
    printf '%b\n' ""
    printf '%s' "Digite os números desejados: "
    read -r -a selections

    if process_selections "${selections[@]}"; then
        process_status=0
    else
        process_status=$?
    fi

    if [[ ${process_status} -eq 10 ]]; then
        return 10
    fi
    if [[ ${process_status} -ne 0 ]]; then
        return 0
    fi

    execute_templates "${templates_dir}"
}

# --- ponto de entrada ---
main() {
    show_menu
}

main "$@"
