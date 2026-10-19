# go-reforged

Correcteur de texte en ligne de commande, écrit en Go. Il reprend la logique de go-reloaded mais est plus exigeant de façon à corriger les lacunes et limites de ce dernier

---

## Utilisation

```
go run . [options] <entrée> <sortie>
```

- `<entrée>` : le fichier à corriger, ou `-` pour lire l'entrée standard ;
- `<sortie>` : le fichier à écrire, ou `-` pour écrire sur la sortie standard. L'entrée et la sortie peuvent être le même fichier.

### Options

| Options | Effets |
| --- | --- |
| `--sentence-case` | majuscule au début de chaque phrase |
| `--articles <fichier>` | fichier d'exceptions pour les articles |
| `--strict` | tout avertissement devient une erreur |
| `--check` | affiche les diagnostics sans rien écrire ; la sortie devient facultative |
| `--workers <n>` | nombre de goroutines de traitement (par défaut : nombre de cœurs) |

### Diagnostics et codes de sortie

Les avertissements et erreurs sont écrits sur la sortie d'erreur, au format :

```
sample.txt:1:8: warning: invalid hexadecimal number "ZZ"
```

Avec l'entrée `-`, le nom affiché est `<stdin>`. La position est celle du modificateur, de la balise ou du guillemet en cause.

| Codes | Effets |
| --- | --- |
| 0 | succès, des avertissements ont pu être affichés |
| 1 | erreur d'utilisation ou de fichier (fichier introuvable, droits…) |
| 2 | erreur de contenu (bloc mal formé, fichier d'articles invalide) |
| 3 | avec `--check` uniquement : des avertissements ont été trouvés|

En cas de code 1 ou 2, le fichier de sortie n'est ni créé ni modifié.

---

## Architecture

```
go-reforged/
├── main.go          mon programme principal
├── lexer/           découpe une ligne en tokens
├── transform/       les règles
├── render/          reconstruit le texte à partir des tokens
├── tools/gen/       générateur de gros fichiers de test
└── testdata/        textes trouvés par le fuzzing, rejoués à chaque go test
```

## Décisions de conception

Les choix que j'ai faits là où le sujet ne précisait rien :

**Lexer**
- Des crochets qui ne forment pas une balise connue restent dans les mots : `see [some note] here` donne les mots `[some` et `note]`.
- Pour un échappement `\(up)`, la colonne retenue est celle de la barre.
- Un modificateur doit avoir exactement la forme `(up)` ou `(up, 2)`, avec une virgule suivie d'un espace. Les formes `( up )` ou `(up,2)` sont du texte ordinaire. Ce choix garantit l'idempotence, vérifiée par le fuzzing.

**Modificateurs**
- Pour une conversion sur plusieurs mots, chaque mot est traité indépendamment : `FF ZZ (hex, 2)` donne `255 ZZ (hex, 2)`. Le modificateur est conservé dès qu'un mot n'a pas pu être converti.

**Articles**
- Le « mot suivant » est le prochain mot, en ignorant la ponctuation et les guillemets : `a , apple` et `a 'apple'` deviennent `an , apple` et `an 'apple'`.
- Seule la casse de la première lettre de l'article est conservée : `A ORANGE` devient `An ORANGE`.
- Un article n'est pas modifié quand un modificateur se trouve entre lui et le mot suivant, comme dans `An (hex) B`.

**Fichier d'exceptions**
- Les lignes vides sont ignorées.
- Une ligne valide contient exactement deux éléments, le second étant `a` ou `an`, quelle que soit sa casse. Toute autre ligne est une erreur.

**Parenthèses**
- Seules les paires sont collées. Une parenthèse sans partenaire reste telle quelle, avec un avertissement.
- Une paire n'est pas collée si le résultat formerait un modificateur, comme `( up )`.

**Majuscule en début de phrase**
- Seuls `.`, `!` et `?` isolés terminent une phrase. Les groupes comme `...` ou `?!` n'en terminent pas : `really?! yes` donne `Really?! yes`.

---

## Limites connues

- **Idempotence et échappement.** Le sujet impose que `\(up)` devienne `(up)`. Retraité, ce `(up)` est alors un modificateur : un texte contenant un échappement n'est donc pas idempotent. Le test de fuzzing exclut ces textes

- **Idempotence et erreurs.** Un texte qui produit une erreur ne donne aucun résultat, il n'y a donc rien à retraiter. Le fuzzing exclut aussi ces textes

- **`--sentence-case` et workers.** La règle de majuscule retient la fin de la ligne précédente : elle ne peut pas être appliquée en parallèle. Avec cette option, le traitement utilise un seul worker

- **Ponctuation dans un mot.** Un point entre deux chiffres est traité comme une ponctuation : `3.5` devient `3. 5`

- **Sortie standard et erreurs.** Avec la sortie `-`, les lignes sont affichées au fil de l'eau. Si une erreur apparaît en fin de fichier, les lignes précédentes ont déjà été affichées.

- **Ramasse-miettes.** Le traitement alloue beaucoup de mémoire temporaire, ce qui freinait les workers. Le ramasse-miettes est réglé avec `debug.SetGCPercent(400)` pour passer moins souvent ; la mémoire utilisée reste limitée à un paquet de lignes.

- **Options invalides.** Une option inconnue ou une valeur invalide, comme `--workers abc`, est gérée par le package `flag`, qui termine avec le code 2 au lieu du code 1.

- **Modifications de texte.** Une apostrophe en bord de mot est traitée comme un guillemet : sans partenaire, elle est séparée du mot, `students'` donnant `students '`. Une barre qui n'échappe pas un modificateur reste un mot séparé : `\(hello)` donne `\ (hello)`. Un modificateur mal formé comme `(up,2)` reste du texte, avec ses parenthèses séparées.

---

## Tests et performances

```
go test ./...
```

**Couverture** (au-dessus de 80 %) :

```
go test ./... "-coverprofile=couverture.out"
go tool cover "-func=couverture.out"
```

**Fuzzing** : vérifie que le traitement ne plante jamais, et qu'un texte corrigé n'est plus modifié par un second passage.

```
go test "-fuzz=FuzzTraitement" "-fuzztime=30s"
```

Il a notamment permis de découvrir et corriger : des parenthèses recollées qui fabriquaient un modificateur, et un article modifié sous un modificateur conservé.

**Benchmark** :

```
go test "-bench=TraitementLigne" "-benchmem"
```

**Générateur et mesures** :

```
go run ./tools/gen 100MB big.txt
go build -o correcteur.exe .
```

Mesures sur un fichier de 100 Mo sans avertissements, avec un Intel Core Ultra 7 356H :

| Workers | Durée |
| --- | --- |
| 1 | 7,8 s |
| 4 | 3,7 s |

Une ligne unique de 10 Mo est traitée en moins de deux secondes.