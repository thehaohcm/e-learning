# E-Learn: save words (browser extension)

Select an English word on any page → **right-click → Add “…” to E-Learn notes**, or press **Alt+Shift+N**.
The word is saved to *My notes* on the website together with the sentence it was in (and, if allowed, the
site name, e.g. `bbc.com`). On the website, open 📝 → **Explain** (AI meaning in that context) → **Add to Study**.

Works in Chrome, Edge, Brave, Cốc Cốc (Manifest V3).

## Install
1. Open `chrome://extensions` (Edge: `edge://extensions`) and turn on **Developer mode**.
2. **Load unpacked** → choose this `extension` folder.
3. On the website: **Settings → Browser extension → Create key**, copy the key.
4. The extension options open after installing (or right-click the icon → *Options*): set the website address
   (default `https://elearn.tradesign.io.vn`), paste the key, **Save & test** → “Connected as …”.

Change the shortcut at `chrome://extensions/shortcuts`.

## Privacy
Only the selected word, its sentence and (optional) the site name are sent — never the full page address or page content.
The key can only add/read your notes; remove it any time in Settings (the extension stops working at once).

## Files
- `background.js` — right-click menu, shortcut, reads the selection + sentence (only in the tab you act on: `activeTab`), saves it.
- `popup.html/js` — add a word by typing, recently saved words.
- `options.html/js` — website address, key, site-name option.
- API used: `GET /api/ext/me`, `GET /api/ext/notes`, `POST /api/ext/notes` with `Authorization: Bearer ext_…`.
