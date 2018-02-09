package utl

const (
	extVsmime = `{
    "js":    "application/vnd.google-apps.script+json",
    "gs":    "application/vnd.google-apps.script+json",
    "gas":   "application/vnd.google-apps.script+json",
    "csv":   "text/csv",
    "htm":   "text/html",
    "html":  "text/html",
    "xbm":   "text/html",
    "shtml": "text/html",
    "shtm":  "text/html",
    "txt":   "text/plain",
    "text":  "text/plain",
    "json":  "application/json",
    "doc":   "application/msword",
    "xls":   "application/vnd.ms-excel",
    "ppt":   "application/vnd.ms-powerpoint",
    "docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    "xlsx":  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    "pptx":  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    "pdf":   "application/pdf",
    "ps":    "application/postscript",
    "eps":   "application/postscript",
    "gif":   "image/gif",
    "png":   "image/png",
    "svg":   "image/svg+xml",
    "jpg":   "image/jpeg",
    "jpeg":  "image/jpeg",
    "bmp":   "image/bmp",
    "ico":   "image/x-icon",
    "tif":   "image/tiff",
    "tiff":  "image/tiff",
    "mp3":   "audio/mp3",
    "wav":   "audio/wav",
    "mp4":   "video/mp4",
    "zip":   "application/zip"
    }`

	googlemimetypes = `{
    "importFormats": {
        "application/x-vnd.oasis.opendocument.presentation": [
            "application/vnd.google-apps.presentation"
        ],
        "text/tab-separated-values": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "image/jpeg": [
            "image/jpeg"
        ],
        "image/bmp": [
            "image/bmp"
        ],
        "image/gif": [
            "image/gif"
        ],
        "application/vnd.ms-excel.sheet.macroenabled.12": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "application/vnd.openxmlformats-officedocument.wordprocessingml.template": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.ms-powerpoint.presentation.macroenabled.12": [
            "application/vnd.google-apps.presentation"
        ],
        "application/vnd.ms-word.template.macroenabled.12": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [
            "application/vnd.google-apps.document"
        ],
        "image/pjpeg": [
            "image/pjpeg"
        ],
        "application/vnd.google-apps.script+text/plain": [
            "application/vnd.google-apps.script"
        ],
        "application/vnd.ms-excel": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "application/vnd.sun.xml.writer": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.ms-word.document.macroenabled.12": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.ms-powerpoint.slideshow.macroenabled.12": [
            "application/vnd.google-apps.presentation"
        ],
        "text/rtf": [
            "application/vnd.google-apps.document"
        ],
        "text/plain": [
            "text/plain"
        ],
        "application/vnd.oasis.opendocument.spreadsheet": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "application/x-vnd.oasis.opendocument.spreadsheet": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "image/png": [
            "image/png"
        ],
        "application/x-vnd.oasis.opendocument.text": [
            "application/vnd.google-apps.document"
        ],
        "application/msword": [
            "application/vnd.google-apps.document"
        ],
        "application/pdf": [
            "application/pdf"
        ],
        "application/json": [
            "application/json"
        ],
        "application/x-msmetafile": [
            "application/vnd.google-apps.drawing"
        ],
        "application/vnd.openxmlformats-officedocument.spreadsheetml.template": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "application/vnd.ms-powerpoint": [
            "application/vnd.google-apps.presentation"
        ],
        "application/vnd.ms-excel.template.macroenabled.12": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "image/x-bmp": [
            "image/x-bmp"
        ],
        "application/rtf": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.openxmlformats-officedocument.presentationml.template": [
            "application/vnd.google-apps.presentation"
        ],
        "image/x-png": [
            "image/x-png"
        ],
        "text/html": [
            "text/html"
        ],
        "application/vnd.oasis.opendocument.text": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.openxmlformats-officedocument.presentationml.presentation": [
            "application/vnd.google-apps.presentation"
        ],
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "application/vnd.google-apps.script+json": [
            "application/vnd.google-apps.script"
        ],
        "application/vnd.openxmlformats-officedocument.presentationml.slideshow": [
            "application/vnd.google-apps.presentation"
        ],
        "application/vnd.ms-powerpoint.template.macroenabled.12": [
            "application/vnd.google-apps.presentation"
        ],
        "text/csv": [
            "text/csv"
        ],
        "application/vnd.oasis.opendocument.presentation": [
            "application/vnd.google-apps.presentation"
        ],
        "image/jpg": [
            "image/jpg"
        ],
        "text/richtext": [
            "application/vnd.google-apps.document"
        ],
        "application/postscript": [
            "application/postscript"
        ],
        "image/x-icon": [
            "image/x-icon"
        ],
        "image/tiff": [
            "image/tiff"
        ],
        "video/mp4": [
            "video/mp4"
        ],
        "application/zip": [
            "application/zip"
        ],
        "image/svg+xml": [
            "image/svg+xml"
        ],
        "application/postscript": [
            "application/postscript"
        ]
    },
    "exportFormats": {
        "application/vnd.google-apps.form": [
            "application/zip"
        ],
        "application/vnd.google-apps.document": [
            "application/rtf",
            "application/vnd.oasis.opendocument.text",
            "text/html",
            "application/pdf",
            "application/epub+zip",
            "application/zip",
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
            "text/plain"
        ],
        "application/vnd.google-apps.drawing": [
            "image/svg+xml",
            "image/png",
            "application/pdf",
            "image/jpeg"
        ],
        "application/vnd.google-apps.spreadsheet": [
            "application/x-vnd.oasis.opendocument.spreadsheet",
            "text/tab-separated-values",
            "application/pdf",
            "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
            "text/csv",
            "application/zip",
            "application/vnd.oasis.opendocument.spreadsheet"
        ],
        "application/vnd.google-apps.script": [
            "application/vnd.google-apps.script+json"
        ],
        "application/vnd.google-apps.presentation": [
            "application/vnd.oasis.opendocument.presentation",
            "application/pdf",
            "application/vnd.openxmlformats-officedocument.presentationml.presentation",
            "text/plain"
        ]
    }
    }`

	defaultformat = `{
        "application/vnd.google-apps.form": "application/zip",
        "application/vnd.google-apps.document": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
        "application/vnd.google-apps.drawing": "image/png",
        "application/vnd.google-apps.spreadsheet": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "application/vnd.google-apps.script": "application/vnd.google-apps.script+json",
        "application/vnd.google-apps.presentation": "application/vnd.openxmlformats-officedocument.presentationml.presentation"
    }`
)
