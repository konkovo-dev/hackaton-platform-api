package mappers

import (
	commonv1 "github.com/belikoooova/hackaton-platform-api/api/common/v1"
	identityv1 "github.com/belikoooova/hackaton-platform-api/api/identity/v1"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/domain/entity"
	"github.com/belikoooova/hackaton-platform-api/internal/identity-service/usecase/skills"
)

func CatalogSkillsToProto(catalogSkills []*entity.CatalogSkill, customSkills []*entity.CustomSkill) []*identityv1.Skill {
	skills := make([]*identityv1.Skill, 0, len(catalogSkills)+len(customSkills))

	for _, catalogSkill := range catalogSkills {
		skills = append(skills, &identityv1.Skill{
			Kind: &identityv1.Skill_Catalog{
				Catalog: &identityv1.CatalogSkill{
					Id:   catalogSkill.ID.String(),
					Name: catalogSkill.Name,
				},
			},
		})
	}

	for _, customSkill := range customSkills {
		skills = append(skills, &identityv1.Skill{
			Kind: &identityv1.Skill_Custom{
				Custom: &identityv1.CustomSkill{
					Name: customSkill.Name,
				},
			},
		})
	}

	return skills
}

func ListSkillCatalogOutToResponse(out *skills.ListSkillCatalogOut) *identityv1.ListSkillCatalogResponse {
	catalogSkills := make([]*identityv1.CatalogSkill, 0, len(out.Skills))

	for _, skill := range out.Skills {
		catalogSkills = append(catalogSkills, &identityv1.CatalogSkill{
			Id:   skill.ID.String(),
			Name: skill.Name,
		})
	}

	return &identityv1.ListSkillCatalogResponse{
		Skills: catalogSkills,
		Page: &commonv1.PageResponse{
			NextPageToken: out.NextPageToken,
		},
	}
}
