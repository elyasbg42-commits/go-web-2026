// web/static/js/filtre.js
// Recherche instantanée : à chaque frappe, on demande du JSON à NOTRE serveur Go
// (jamais à la mini API) et on reconstruit la liste sans recharger la page.
// Sans JavaScript, le formulaire GET marche toujours : le JS vient par-dessus.
const input = document.querySelector("#recherche");
const select = document.querySelector("#genre");
const liste = document.querySelector("#liste");

async function filtrer() {
    const params = new URLSearchParams({ q: input.value, genre: select.value });

    let games;
    try {
        const resp = await fetch("/api/jeux?" + params);
        if (!resp.ok) throw new Error("statut " + resp.status);
        games = await resp.json();
    } catch (err) {
        console.error("Filtre impossible :", err);
        return; // on laisse la liste telle quelle
    }

    liste.innerHTML = "";
    if (games.length === 0) {
        const li = document.createElement("li");
        li.textContent = "Aucun jeu ne correspond.";
        liste.append(li);
        return;
    }
    for (const g of games) {
        const li = document.createElement("li");

        const a = document.createElement("a");
        a.href = "/jeux/" + g.id;
        a.textContent = g.name; // textContent : le texte s'affiche tel quel

        const annee = document.createElement("span");
        annee.className = "annee";
        annee.textContent = g.year;

        li.append(a, " ", annee);
        liste.append(li);
    }
}

input.addEventListener("input", filtrer);
select.addEventListener("change", filtrer);