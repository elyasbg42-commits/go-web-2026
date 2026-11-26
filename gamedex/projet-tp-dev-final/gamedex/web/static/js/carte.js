// web/static/js/carte.js
// `tournois` est défini juste avant, par le template (liste d'objets JSON).
const carte = L.map("carte").setView([46.6, 2.4], 4);

L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
    attribution: "© OpenStreetMap",
    maxZoom: 19,
}).addTo(carte);

const points = [];
for (const t of tournois) {
    // Le texte vient d'une API : on le place avec textContent, jamais en HTML brut.
    const bulle = document.createElement("div");
    const titre = document.createElement("strong");
    titre.textContent = t.name;
    bulle.append(titre, document.createElement("br"), t.city + " · " + t.date);

    L.marker([t.lat, t.lng]).addTo(carte).bindPopup(bulle);
    points.push([t.lat, t.lng]);
}

// Cadrer la carte sur tous les tournois du jeu.
if (points.length > 0) {
    carte.fitBounds(points, { padding: [40, 40], maxZoom: 7 });
}