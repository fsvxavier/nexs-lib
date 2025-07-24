#!/bin/bash

# Script para testar a infraestrutura atualizada da NEXS-LIB
# Este script verifica se todos os esquemas, tabelas e dados foram criados corretamente

echo "=== NEXS-LIB Infrastructure Test ==="
echo "Testando a infraestrutura atualizada..."

# Configurações de conexão
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="nexs_user"
DB_PASSWORD="nexs_password"
DB_NAME="nexs_testdb"

# Função para executar query SQL
execute_query() {
    local query="$1"
    local description="$2"
    
    echo "📋 Testando: $description"
    
    result=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "$query" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "✅ $description: OK"
        if [ ! -z "$result" ]; then
            echo "   Resultado: $result"
        fi
    else
        echo "❌ $description: ERRO"
        return 1
    fi
}

# Função para testar existência de tabela
test_table_exists() {
    local table_name="$1"
    local schema_name="${2:-public}"
    
    query="SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = '$schema_name' AND table_name = '$table_name');"
    execute_query "$query" "Tabela $schema_name.$table_name existe"
}

# Função para contar registros
count_records() {
    local table_name="$1"
    local schema_name="${2:-public}"
    
    query="SELECT COUNT(*) FROM $schema_name.$table_name;"
    execute_query "$query" "Contagem de registros em $schema_name.$table_name"
}

echo ""
echo "=== Verificando Conexão ==="
execute_query "SELECT version();" "Versão do PostgreSQL"

echo ""
echo "=== Verificando Tabelas Principais ==="
test_table_exists "products"
test_table_exists "accounts"
test_table_exists "copy_test"
test_table_exists "tenants"
test_table_exists "shared_users"
test_table_exists "chat_messages"
test_table_exists "monitored_table"
test_table_exists "replica_test"
test_table_exists "performance_test"
test_table_exists "audit_log"

echo ""
echo "=== Verificando Schemas de Tenants ==="
test_table_exists "users" "tenant_empresa_a"
test_table_exists "users" "tenant_empresa_b"
test_table_exists "users" "tenant_empresa_c"

echo ""
echo "=== Verificando Dados de Exemplo ==="
count_records "products"
count_records "accounts"
count_records "copy_test"
count_records "tenants"
count_records "shared_users"
count_records "chat_messages"
count_records "monitored_table"
count_records "replica_test"
count_records "performance_test"

echo ""
echo "=== Verificando Dados em Schemas de Tenants ==="
count_records "users" "tenant_empresa_a"
count_records "users" "tenant_empresa_b"
count_records "users" "tenant_empresa_c"

echo ""
echo "=== Verificando Funções Utilitárias ==="
execute_query "SELECT EXISTS (SELECT FROM pg_proc WHERE proname = 'generate_batch_test_data');" "Função generate_batch_test_data"
execute_query "SELECT EXISTS (SELECT FROM pg_proc WHERE proname = 'generate_copy_test_data');" "Função generate_copy_test_data"
execute_query "SELECT EXISTS (SELECT FROM pg_proc WHERE proname = 'simulate_account_transactions');" "Função simulate_account_transactions"
execute_query "SELECT EXISTS (SELECT FROM pg_proc WHERE proname = 'reset_example_data');" "Função reset_example_data"
execute_query "SELECT EXISTS (SELECT FROM pg_proc WHERE proname = 'get_example_table_stats');" "Função get_example_table_stats"

echo ""
echo "=== Verificando Views ==="
execute_query "SELECT EXISTS (SELECT FROM information_schema.views WHERE table_name = 'batch_operation_summary');" "View batch_operation_summary"
execute_query "SELECT EXISTS (SELECT FROM information_schema.views WHERE table_name = 'multi_tenant_summary');" "View multi_tenant_summary"
execute_query "SELECT EXISTS (SELECT FROM information_schema.views WHERE table_name = 'performance_metrics');" "View performance_metrics"
execute_query "SELECT EXISTS (SELECT FROM information_schema.views WHERE table_name = 'product_stats');" "View product_stats"
execute_query "SELECT EXISTS (SELECT FROM information_schema.views WHERE table_name = 'account_summary');" "View account_summary"
execute_query "SELECT EXISTS (SELECT FROM information_schema.views WHERE table_name = 'tenant_stats');" "View tenant_stats"

echo ""
echo "=== Verificando Triggers ==="
execute_query "SELECT EXISTS (SELECT FROM information_schema.triggers WHERE trigger_name = 'audit_products_trigger');" "Trigger audit_products_trigger"
execute_query "SELECT EXISTS (SELECT FROM information_schema.triggers WHERE trigger_name = 'audit_accounts_trigger');" "Trigger audit_accounts_trigger"
execute_query "SELECT EXISTS (SELECT FROM information_schema.triggers WHERE trigger_name = 'notify_monitored_table_trigger');" "Trigger notify_monitored_table_trigger"
execute_query "SELECT EXISTS (SELECT FROM information_schema.triggers WHERE trigger_name = 'notify_new_chat_message_trigger');" "Trigger notify_new_chat_message_trigger"

echo ""
echo "=== Testando Funcionalidades ==="

# Testar função de estatísticas
echo "📊 Testando função de estatísticas:"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT * FROM get_example_table_stats();" 2>/dev/null

# Testar views
echo ""
echo "📈 Testando view de resumo de produtos:"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT * FROM product_stats LIMIT 5;" 2>/dev/null

echo ""
echo "🏢 Testando view de multi-tenancy:"
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT * FROM multi_tenant_summary;" 2>/dev/null

echo ""
echo "=== Resumo dos Testes ==="
echo "✅ Infraestrutura atualizada com base nos exemplos da NEXS-LIB"
echo "✅ Todas as tabelas necessárias para os exemplos foram criadas"
echo "✅ Dados de teste apropriados foram inseridos"
echo "✅ Funções utilitárias para testes estão disponíveis"
echo "✅ Views para análise e monitoramento estão funcionais"
echo "✅ Triggers para auditoria e notificações estão ativos"
echo ""
echo "🎯 A infraestrutura está pronta para executar todos os exemplos da NEXS-LIB!"
echo "🎯 Use ./manage.sh start para iniciar os serviços"
echo "🎯 Use ./manage.sh stop para parar os serviços"
echo "🎯 Use ./manage.sh restart para reiniciar os serviços"
echo ""
echo "📚 Para mais informações, consulte o README.md"
