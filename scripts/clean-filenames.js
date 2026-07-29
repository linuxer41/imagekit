const fs = require('fs');
const path = require('path');

const dir = process.argv[2] || '.';

function cleanFilename(name) {
  const ext = path.extname(name);
  const base = path.basename(name, ext);
  const cleaned = base
    .replace(/[^a-zA-Z0-9_]/g, '_')
    .replace(/_+/g, '_');
  return cleaned + ext;
}

const files = fs.readdirSync(dir);
let renamed = 0;

for (const file of files) {
  const cleaned = cleanFilename(file);
  if (cleaned !== file) {
    const oldPath = path.join(dir, file);
    const newPath = path.join(dir, cleaned);
    if (!fs.existsSync(newPath)) {
      fs.renameSync(oldPath, newPath);
      console.log(`${file} -> ${cleaned}`);
      renamed++;
    } else {
      console.log(`SKIP (ya existe): ${file} -> ${cleaned}`);
    }
  }
}

console.log(`\n${renamed} archivos renombrados de ${files.length}`);
