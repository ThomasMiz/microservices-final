# Créditos
Esta guía fue originalmente desarrollada por Franco Rupnik para la materia 73.40 - Arquitectura de Microservicios en el Instituto Tecnológico de Buenos Aires. Nosotros realizamo las siguientes modificaciones:
- Reemplazamos los valores en `locals.tf`.
- Agregamos al módulo `k8s_ec2` una variable de entrada para permitir especificar el rol IAM a asignarle a las instancias EC2.
- Actualizamos `main.tf` para buscar el rol existente `LabRole` y pasarlo a las instancias EC2.
- Agregamos un módulo `s3-front` que crea un bucket S3 público y carga el front-end (que se espera que esté en `${path.root}/../../chotel-front`) al bucket.
    - Agregamos un output `frontend_s3_static_website_url` con el _static website endpoint_ de este bucket.
- Agregamos una porción de chocotorta (te la tenés que imaginar).

# Guía Rápida para Desplegar tu Cluster Kubernetes en AWS

## **Dependencias Previas**
Asegúrate de tener instalados:
- **Terraform**
- **AWS CLI**
- **Kubectl** (opcional, para conectarte al cluster)
- **jq** (para parsear JSON en bash y obtener el output de Terraform)

> **Recorda:** Configura AWS CLI con un usuario que tenga permisos para crear recursos (EC2, VPC, IAM, etc).

---

## **Generar Clave SSH**
Para conectarte a las instancias EC2, genera una clave SSH:
```bash
ssh-keygen -t rsa -b 4096 -C "tuemail_de_preferencia@gmail.com" -f ./nombre_de_tu_clave
```
Esto creará:
- `nombre_de_tu_clave` (clave privada)
- `nombre_de_tu_clave.pub` (clave pública para AWS)

Cambia los permisos de la clave privada:
```bash
chmod 400 ./nombre_de_tu_clave
```
Si hay problemas con los permisos, asegurate de estar en un entorno Unix (Linux/Mac) porque Windows no deja aplicar bien los permisos.
Podes usar WSL2, pero necesitas no podes estar en la carpeta de Windows, tenes que estar en una carpeta dentro de WSL2 (ej: `/home/tu_usuario/...`).

---

## **Creación del Cluster**
1. **Inicializa Terraform:**
   ```bash
   terraform init
   ```
2. **Configura variables en `locals.tf`:**
   - Cambia `key_file_name` por el path de tu clave pública (`.pub`).
   - Ajusta `region` y `profile` si deseas usar algo distinto a los valores default.
   - Modifica `worker_count` para definir la cantidad de nodos workers.
3. **Aplica la infraestructura:**
   ```bash
   terraform apply -auto-approve
   ```
   > ⏳ Espera ~5 minutos para que las instancias estén listas y el cluster operativo.

---

## **Conectar los Nodos Workers**
Ejecuta:
```bash
./setup_cluster.sh <path_a_tu_clave_privada>
```
- `<path_a_tu_clave_privada>`: tu archivo de clave privada (sin extensión).
- El script pedirá confirmación `yes` al conectarse por primera vez a cada nodo.

---

## **Obtener Configuración de Kubectl**
Ejecuta:
```bash
./connect_kubectl.sh <path_a_tu_clave_privada>
```
Esto copiará el archivo de configuración a `~/.kube/config` (haz backup si ya tenes uno y no queres perderlo).

Verifica el estado del cluster:
```bash
kubectl get nodes
```
Deberías ver los nodos en estado `Ready`.

> **💡 TIP:** Se recomienda **reiniciar manualmente las instancias EC2** desde la consola de AWS tras la instalación para evitar problemas de conectividad entre pods o servicios.

---

## **Destrucción del Cluster**
Para eliminar todos los recursos:
```bash
terraform destroy -auto-approve
```
Esto borrará todo lo creado por Terraform en AWS.

---

## **Visualización y Manejo del Cluster**
¿No queres usar la terminal o extensiones de VSCode? Te recomiendo usar el IDE [Lens](https://k8slens.dev/):
- Gratuito y Open Source.
- Detecta automáticamente tu archivo `~/.kube/config`.
- Te facilita mucho la gestión visual de tu cluster dejandote ver nodos, pods, servicios, etc.
- Podes rapidamente matar pods, volumenes, volume claims, generar tuneles, revisar logs, etc.

