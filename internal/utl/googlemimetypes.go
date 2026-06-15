package utl

const (
	extVsmime = `{
    "js":        "application/vnd.google-apps.script+json",
    "gs":        "application/vnd.google-apps.script+json",
    "gas":       "application/vnd.google-apps.script+json",
    "csv":       "text/csv",
    "tsv":       "text/tab-separated-values",
    "htm":       "text/html",
    "html":      "text/html",
    "txt":       "text/plain",
    "text":      "text/plain",
    "json":      "application/json",
    "doc":       "application/msword",
    "xls":       "application/vnd.ms-excel",
    "ppt":       "application/vnd.ms-powerpoint",
    "docx":      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    "xlsx":      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    "pptx":      "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    "pdf":       "application/pdf",
    "ps":        "application/postscript",
    "eps":       "application/postscript",
    "gif":       "image/gif",
    "png":       "image/png",
    "svg":       "image/svg+xml",
    "jpg":       "image/jpeg",
    "jpeg":      "image/jpeg",
    "bmp":       "image/jpeg",
    "ico":       "image/x-icon",
    "tif":       "image/tiff",
    "tiff":      "image/tiff",
    "mp3":       "audio/mp3",
    "wav":       "audio/wav",
    "mp4":       "video/mp4",
    "ogg":       "video/ogg",
    "mov":       "video/quicktime",
    "webm":      "video/webm",
    "zip":       "application/zip",
    "py":        "text/x-python",
    "md":        "text/markdown",
    "markdown":  "text/markdown",
    "rtf":       "application/rtf",
    "odt":       "application/vnd.oasis.opendocument.text",
    "ods":       "application/vnd.oasis.opendocument.spreadsheet",
    "odp":       "application/vnd.oasis.opendocument.presentation",
    "epub":      "application/epub+zip"
    }`

	googlemimetypes = `{
    "importFormats": {
        "application/x-vnd.oasis.opendocument.presentation": [
            "application/vnd.google-apps.presentation"
        ],
        "text/tab-separated-values": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "image/gif": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.ms-excel.sheet.macroenabled.12": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "application/vnd.openxmlformats-officedocument.wordprocessingml.template": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.ms-word.template.macroenabled.12": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [
            "application/vnd.google-apps.document"
        ],
        "video/ogg": [
            "application/vnd.google-apps.vid"
        ],
        "application/vnd.ms-excel": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "text/rtf": [
            "application/vnd.google-apps.document"
        ],
        "application/x-vnd.oasis.opendocument.text": [
            "application/vnd.google-apps.document"
        ],
        "application/msword": [
            "application/vnd.google-apps.document"
        ],
        "application/pdf": [
            "application/vnd.google-apps.document"
        ],
        "application/x-msmetafile": [
            "application/vnd.google-apps.drawing"
        ],
        "text/markdown": [
            "application/vnd.google-apps.document"
        ],
        "image/x-bmp": [
            "application/vnd.google-apps.document"
        ],
        "application/rtf": [
            "application/vnd.google-apps.document"
        ],
        "text/html": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.oasis.opendocument.text": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.openxmlformats-officedocument.presentationml.presentation": [
            "application/vnd.google-apps.presentation"
        ],
        "text/csv": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "application/vnd.oasis.opendocument.presentation": [
            "application/vnd.google-apps.presentation"
        ],
        "image/jpg": [
            "application/vnd.google-apps.document"
        ],
        "video/quicktime": [
            "application/vnd.google-apps.vid"
        ],
        "text/richtext": [
            "application/vnd.google-apps.document"
        ],
        "video/mp4": [
            "application/vnd.google-apps.vid"
        ],
        "video/webm": [
            "application/vnd.google-apps.vid"
        ],
        "image/jpeg": [
            "application/vnd.google-apps.document"
        ],
        "image/bmp": [
            "application/vnd.google-apps.document"
        ],
        "text/x-markdown": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.ms-powerpoint.presentation.macroenabled.12": [
            "application/vnd.google-apps.presentation"
        ],
        "text/comma-separated-values": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "image/pjpeg": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.google-apps.script+text/plain": [
            "application/vnd.google-apps.script"
        ],
        "application/vnd.ms-word.document.macroenabled.12": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.ms-powerpoint.slideshow.macroenabled.12": [
            "application/vnd.google-apps.presentation"
        ],
        "text/plain": [
            "application/vnd.google-apps.document"
        ],
        "application/vnd.oasis.opendocument.spreadsheet": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "application/x-vnd.oasis.opendocument.spreadsheet": [
            "application/vnd.google-apps.spreadsheet"
        ],
        "image/png": [
            "application/vnd.google-apps.document"
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
        "application/vnd.openxmlformats-officedocument.presentationml.template": [
            "application/vnd.google-apps.presentation"
        ],
        "image/x-png": [
            "application/vnd.google-apps.document"
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
        ]
    },
    "exportFormats": {
        "application/vnd.google-apps.document": [
            "application/rtf",
            "application/vnd.oasis.opendocument.text",
            "text/html",
            "application/pdf",
            "text/x-markdown",
            "text/markdown",
            "application/epub+zip",
            "application/zip",
            "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
            "text/plain"
        ],
        "application/vnd.google-apps.vid": [
            "video/mp4"
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
        ],
        "application/vnd.google-apps.form": [
            "application/zip"
        ],
        "application/vnd.google-apps.drawing": [
            "image/svg+xml",
            "image/png",
            "application/pdf",
            "image/jpeg"
        ],
        "application/vnd.google-apps.site": [
            "text/plain"
        ],
        "application/vnd.google-apps.mail-layout": [
            "text/plain"
        ],
        "application/vnd.google-apps.pix": [
            "image/jpeg",
            "image/png"
        ]
    }
}`

	defaultformat = `{
        "application/vnd.google-apps.form": "application/zip",
        "application/vnd.google-apps.document": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
        "application/vnd.google-apps.drawing": "application/pdf",
        "application/vnd.google-apps.spreadsheet": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "application/vnd.google-apps.script": "application/vnd.google-apps.script+json",
        "application/vnd.google-apps.presentation": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
        "application/vnd.google-apps.site": "text/plain",
        "application/vnd.google-apps.jam": "application/pdf",
        "application/vnd.google-apps.vid": "video/mp4",
        "application/vnd.google-apps.mail-layout": "text/plain",
        "application/vnd.google-apps.pix": "image/png"
    }`

	mimeVsEx = `{
        "application/vnd.google-apps.script": ".gs",
        "application/vnd.google-apps.script+json": ".gs",
        "text/csv": ".csv",
        "text/tab-separated-values": ".tsv",
        "text/html": ".html",
        "text/plain": ".txt",
        "application/json": ".json",
        "application/msword": ".doc",
        "application/vnd.ms-excel": ".xls",
        "application/vnd.ms-powerpoint": ".ppt",
        "application/vnd.google-apps.document": ".docx",
        "application/vnd.google-apps.spreadsheet": ".xlsx",
        "application/vnd.google-apps.presentation": ".pptx",
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
        "application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
        "application/pdf": ".pdf",
        "application/postscript": ".ps",
        "image/gif": ".gif",
        "image/png": ".png",
        "image/svg+xml": ".svg",
        "image/jpeg": ".jpg",
        "image/bmp": ".bmp",
        "image/x-icon": ".ico",
        "image/tiff": ".tif",
        "audio/mp3": ".mp3",
        "audio/wav": ".wav",
        "video/mp4": ".mp4",
        "video/ogg": ".ogg",
        "video/quicktime": ".mov",
        "video/webm": ".webm",
        "application/zip": ".zip",
        "text/markdown": ".md",
        "text/x-markdown": ".md",
        "application/rtf": ".rtf",
        "application/vnd.oasis.opendocument.text": ".odt",
        "application/vnd.oasis.opendocument.spreadsheet": ".ods",
        "application/vnd.oasis.opendocument.presentation": ".odp",
        "application/epub+zip": ".epub"
    }`
)
