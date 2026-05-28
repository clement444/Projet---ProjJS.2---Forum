# Forum

Application de forum web développée en Go avec SQLite.

## Prérequis

- [Go 1.22+](https://go.dev/dl/)
- [Docker](https://www.docker.com/products/docker-desktop/)

## Lancement en local

```bash
go run .
```

L'application est accessible sur [http://localhost:8080](http://localhost:8080).

La base de données `forum.db` est créée automatiquement au premier lancement.

## Lancement avec Docker

```bash
docker build -t forum .
docker run -p 8080:8080 forum
```

L'application est accessible sur [http://localhost:8080](http://localhost:8080).

La base de données est stockée dans le conteneur et ne persiste pas entre les redémarrages.

## Fonctionnalités

- Inscription, connexion et déconnexion (sessions avec cookie, expiration 24h)
- Mots de passe hashés avec bcrypt
- Création de posts avec catégories et image optionnelle
- Commentaires sur les posts
- Modification et suppression de ses propres posts et commentaires
- Like / dislike sur les posts et commentaires
- Filtrage des posts par catégorie, par utilisateur connecté, par posts likés
- Pages d'erreur HTML (404, 403, 500)

## Structure du projet

```
forum/
├── main.go
├── go.mod
├── database/
│   └── db.go
├── handlers/
│   ├── auth.go
│   ├── comments.go
│   ├── errors.go
│   ├── likes.go
│   └── posts.go
├── static/
│   ├── css/
│   └── uploads/
├── templates/
│   ├── index.html
│   ├── register.html
│   ├── login.html
│   ├── post.html
│   ├── create_post.html
│   ├── edit_post.html
│   └── error.html
└── Dockerfile
```
