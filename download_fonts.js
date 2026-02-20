const fs = require('fs');
const https = require('https');
const path = require('path');

const cssPath = './public/fonts/google-fonts.css';
let cssContent = fs.readFileSync(cssPath, 'utf8');

const regex = /url\((https:\/\/fonts\.gstatic\.com\/s\/[^\)]+)\)/g;
let match;
const downloads = [];

while ((match = regex.exec(cssContent)) !== null) {
    const url = match[1];
    const filename = url.substring(url.lastIndexOf('/') + 1);
    const localUrl = '/static/fonts/' + filename;

    cssContent = cssContent.replace(url, localUrl);

    downloads.push({ url, filename });
}

fs.writeFileSync(cssPath, cssContent);

downloads.forEach(dl => {
    const dest = path.join('./public/fonts', dl.filename);
    if (!fs.existsSync(dest)) {
        const file = fs.createWriteStream(dest);
        https.get(dl.url, function (response) {
            response.pipe(file);
        });
    }
});

console.log('Fonts downloaded and CSS updated.');
