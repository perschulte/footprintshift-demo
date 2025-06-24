# ⚠️ Dieses Projekt gehört nicht zu EcoCI

## GreenWeb ist ein separates Projekt!

Das `greenweb-api` Verzeichnis sollte aus dem EcoCI-Projekt heraus verschoben werden:

```bash
# Im Terminal ausführen:
cd /Users/perschulte/Documents/dev
mkdir -p greenweb
mv ecoci/greenweb-api greenweb/
cd greenweb/greenweb-api
git init
gh repo create greenweb-api --public
```

## Unterschied der Projekte:

### EcoCI 🌱
- **Was**: Misst CO₂ in CI/CD Pipelines
- **Für wen**: Entwickler, DevOps Teams
- **Use Case**: GitHub Actions, GitLab CI
- **Output**: CO₂ pro Build/Test

### GreenWeb 🌍  
- **Was**: Adaptive Website-Optimierung
- **Für wen**: Website-Betreiber, E-Commerce
- **Use Case**: Dynamische Features basierend auf Strom-Mix
- **Output**: Grünere Websites, Green Hours Pricing