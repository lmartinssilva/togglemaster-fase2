import sys

old_block = '''      - name: Atualiza tag da imagem no repositório GitOps
        run: |
          git clone https://x-access-token:${{ secrets.GITOPS_PAT }}@github.com/lmartinssilva/togglemaster-gitops.git gitops-repo
          cd gitops-repo
          sed -i "s#\\(${{ env.ECR_REPOSITORY }}:\\).*#\\1${{ env.IMAGE_TAG }}#" manifests/deployments.yaml
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add manifests/deployments.yaml
          git commit -m "chore: atualiza ${{ env.ECR_REPOSITORY }} para ${{ env.IMAGE_TAG }}" || echo "Nada para commitar"
          git push'''

new_block = '''      - name: Atualiza tag da imagem no repositório GitOps
        run: |
          for i in 1 2 3 4 5; do
            rm -rf gitops-repo
            git clone https://x-access-token:${{ secrets.GITOPS_PAT }}@github.com/lmartinssilva/togglemaster-gitops.git gitops-repo
            cd gitops-repo
            sed -i "s#\\(${{ env.ECR_REPOSITORY }}:\\).*#\\1${{ env.IMAGE_TAG }}#" manifests/deployments.yaml
            git config user.name "github-actions[bot]"
            git config user.email "github-actions[bot]@users.noreply.github.com"
            git add manifests/deployments.yaml
            if git diff --cached --quiet; then
              echo "Nada para commitar"
              break
            fi
            git commit -m "chore: atualiza ${{ env.ECR_REPOSITORY }} para ${{ env.IMAGE_TAG }}"
            if git push; then
              echo "Push bem-sucedido na tentativa $i"
              break
            fi
            echo "Push falhou (tentativa $i), outro pipeline deve ter chegado primeiro. Tentando de novo..."
            cd ..
            sleep $((RANDOM % 5 + 2))
          done'''

for path in sys.argv[1:]:
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()
    if old_block in content:
        content = content.replace(old_block, new_block)
        with open(path, "w", encoding="utf-8") as f:
            f.write(content)
        print(f"Atualizado: {path}")
    else:
        print(f"AVISO: bloco não encontrado exatamente em {path} — verificar manualmente")
