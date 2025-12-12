-- GetOrganizationAssignments: Lấy danh sách organization IDs được assign cho novel
SELECT organization_id FROM catalog.novel_organization_assignments WHERE novel_id = $1 AND status = 'active'
